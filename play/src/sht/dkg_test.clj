(ns sht.dkg-test
  (:require [next.jdbc :as jdbc]
            [clojure.string :as str]
            [toml.core :as toml]
            [sht.toml-writer :as toml-writer]
            [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [cheshire.core :as json]
            [sht.base64 :as base64]
            [sht.runner :as runner]
            [sht.build :as build]
            [babashka.fs :as fs]
            [sht.play :as play]))

(set! *warn-on-reflection* true)

(defonce play-db-password (or (System/getenv "PLAY_DB_PASSWORD") ""))

(defn decode-epochid
  [^bytes epochid]
  {:epochid/block (.getInt (java.nio.ByteBuffer/wrap (java.util.Arrays/copyOfRange epochid 0 4)))
   :epochid/seq  (.getInt (java.nio.ByteBuffer/wrap (java.util.Arrays/copyOfRange epochid 4 8)))})

(defn- check-query
  [description sys opts query ok-rows?]
  (let [keyper-num (or (:keyper-num opts) (:keyper/num opts))
        db {:dbtype "postgresql"
            :dbname (play/keyper-db keyper-num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        query (if (vector? query)
                query
                (query sys opts))
        description (if (string? description)
                      description
                      (description sys opts))
        rows (jdbc/execute! ds query)]
    {:chk/ok? (boolean (ok-rows? sys opts rows))
     :chk/description description
     :chk/info rows
     :keyper/num keyper-num}))

(defmacro def-check-query
  [id query ok-rows? description]
  `(defmethod runner/check ~id
     [sys# opts#]
     (check-query ~description sys# opts# ~query ~ok-rows?)))

(defn rows-not-empty?
  [sys opts rows]
  (seq rows))

(def-check-query :keyper/tendermint-batch-config-started
  ["select * from tendermint_batch_config"]
  (fn [_ _ rows]
    (empty? (remove :tendermint_batch_config/started rows)))
  "all batch configs should have been started")

(def-check-query :keyper/meta-inf
  ["select * from meta_inf"]
  rows-not-empty?
  "meta_inf table should be filled")

(def-check-query :keyper/eon-exists
  (fn [sys {:keyper/keys [eon]}] ["select * from eons where eon=?" eon])
  rows-not-empty?
  "eon should exist")

(def-check-query :keyper/dkg-success
  ["select * from dkg_result"]
  (fn [_ {:keyper/keys [eon]} rows]
    (->> rows
         (filter (comp #{eon} :dkg_result/eon))
         first
         :dkg_result/success))
  "dkg should finish successfully")

(def-check-query :keyper/dkg-failed
  ["select * from dkg_result where not success"]
  rows-not-empty?
  "dkg should fail")

(def-check-query :keyper/non-zero-activation-block
  ["select * from eons where activation_block_number<=0"]
  (fn [_ _ rows] (empty? rows))
  "activation block number must be positive")

(def-check-query :keyper/keyper-set
  ["select * from keyper_set"]
  (fn [_ {:keyper/keys [expected-count]} rows]
    (= expected-count (count rows)))
  (fn [_ {:keyper/keys [expected-count]}]
    (format "keyper_set table should have %d entries" expected-count)))

(def-check-query :keyper/query
  (fn [sys {:keyper/keys [query]}]
    query)
  (fn [sys {:keyper/keys [expected]} rows]
    (= rows expected))
  (fn [sys {:keyper/keys [description]}]
    description))

(defmethod runner/run ::add-spare-keyper-set
  [sys m]
  (runner/dispatch sys {:run :process/run
                        :process/id :add-spare-keyper-set
                        :process/wait true
                        :process/cmd ["npx" "hardhat" "run" "--network" "localhost" "scripts/add-spare-keyper-set.js"]
                        :process/opts {:dir (str (fs/path play/repo-root "contracts"))}}))

(defmethod runner/run ::configure-keypers
  [sys m]
  (let [deploy-conf (build/sys-deploy-conf sys)
        deploy-conf-path (-> sys :cwd (fs/path "deploy-config-configure-keypers.json") fs/absolutize str)]
    (spit deploy-conf-path (json/encode deploy-conf {:pretty true}))
    (runner/dispatch sys {:run :process/run
                          :process/id :configure-keypers
                          :process/wait true
                          :process/opts {:dir (str (fs/path play/repo-root "contracts"))
                                         :extra-env {"DEPLOY_CONF" deploy-conf-path}}
                          :process/cmd ["npx" "hardhat" "run" "--network" "localhost" "scripts/configure-keypers.js"]})))

(defn- chain-set-ports
  [path seeds n]
  (let [m (toml/read (slurp path) :keywordize true)
        m (-> m
              (assoc-in [:rpc :laddr] (format "tcp://127.0.0.1:%d" (+ 28000 n)))
              (assoc-in [:p2p :laddr] (format "tcp://127.0.0.1:%d" (+ 27000 n)))
              (assoc-in [:p2p :persistent-peers] seeds))]
    (spit path (toml-writer/dump m))))

(defn- sys-chain-config-path
  [sys n & path-elems]
  (str (apply fs/path (:cwd sys) (format "testchain-%d" n) "config", path-elems)))

(defn- init-chains
  [sys]
  (let [conf (:conf sys)
        num-chains (:num-keypers conf)
        num-initial-keypers  (:num-initial-keypers conf)
        genesis-keypers (mapv (comp :eth-address :subcommand/cfg)
                              (take num-initial-keypers (:sys/keypers sys)))
        genesis-args (interleave (repeat "--genesis-keyper") genesis-keypers)
        sys (runner/dispatch sys
                             (mapv (fn [n] {:run :process/run
                                            :process/wait true
                                            :process/id (keyword (str "init-chain-" n))
                                            :process/cmd (apply vector 'rolling-shutter "chain" "init"
                                                                "--root" (format "testchain-%d" n)
                                                                "--blocktime" "1"
                                                                genesis-args)})
                                   (range num-chains)))
        seeds (->> (range num-chains)
                   (mapv (fn [n]
                           (format "%s@127.0.0.1:%d"
                                   (slurp (sys-chain-config-path sys n "node_key.json.id"))
                                   (+ 27000 n)))))]
    (doseq [n (range num-chains)]
      (chain-set-ports (sys-chain-config-path sys n "config.toml")
                       (str/join "," (concat (take n seeds) (drop (inc n) seeds)))
                       n))
    sys))



(defn get-private-validator-key
  "get the private validator key from a priv_validator_key.json map"
  [m]
  (base64/decode (get-in m ["priv_key" "value"])))

(defn seed-from-private-validator-key
  [^bytes pk]
  (java.util.Arrays/copyOfRange pk 0 32))

(defn- read-private-validator-key
  "read the validators private key from the given priv_validator_key.json file"
  [path]
  (let [m (json/decode (slurp path))]
    (get-private-validator-key m)))

(defn- merge-genesis
  "merge multiple genesis.json maps. This returns the first map with the validators key set to the
  concatenation of the validator keys in all given maps"
  [genesis-maps]
  (let [validators (->> genesis-maps
                        (mapcat (fn [m] (get m "validators")))
                        distinct)]
    (assoc (first genesis-maps) "validators" validators)))

(defn- rewrite-genesis
  "rewrite genesis.json. This defines the first num-initial-keypers validators as genesis
  validators and writes the genesis.json files"
  ([cwd num-keypers num-initial-keypers]
   (let [paths (mapv (fn [n]
                       (-> cwd (fs/path (format "testchain-%d" n) "config" "genesis.json") str))
                     (range num-keypers))
         gens (mapv (fn [p] (-> p slurp json/decode)) (take num-initial-keypers paths))
         genesis (merge-genesis gens)
         genesis-str (json/encode genesis {:pretty true})]
     (doseq [p paths]
       (spit p genesis-str))))
  ([{:keys [conf cwd] :as sys}]
   (let [num-keypers (:num-keypers conf)
         num-initial-keypers (:num-initial-keypers conf)]
     (rewrite-genesis cwd num-keypers num-initial-keypers))))

(defn- format-hex
 [^bytes bs]
  (.formatHex (java.util.HexFormat/of) bs))
;; (format-hex (byte-array 5 [1 3 32 128 255]))

(defn- set-seeds
  "Set the ValidatorSeed value in the keyper's configs. They need to be written out with
  build/sys-write-config-files afterwards."
  [sys]
  (let [keypers (:sys/keypers sys)
        seeds (mapv (fn [n]
                      (let [privkey (read-private-validator-key
                                     (sys-chain-config-path sys n "priv_validator_key.json"))
                            seed (seed-from-private-validator-key privkey)
                            seed-str (format-hex seed)]
                        seed-str))
                    (range (count keypers)))
        keypers (mapv (fn [k seed n]
                        (-> k
                            (assoc-in [:subcommand/toml-edits "ValidatorSeed"] seed)
                            (assoc-in [:subcommand/toml-edits "ShuttermintURL"]
                                      (format "http://localhost:%d" (+ 28000 n)))))
                      keypers
                      seeds
                      (range (count keypers)))]
    (assoc sys :sys/keypers keypers)))

(defmethod runner/run :chain/run-chains
  [sys m]
  (let [num-chains (-> sys :conf :num-keypers)
        sys (init-chains sys)
        sys (set-seeds sys)]
    (rewrite-genesis sys)
    (build/sys-write-config-files sys)
    (runner/dispatch sys
                     (mapv (fn [n]
                             {:run :process/run
                              :process/wait false
                              :process/id (keyword (str "chain-" n))
                              :process/cmd ['rolling-shutter "chain" "--config"
                                            (format "testchain-%d/config/config.toml" n)]})
                           (range num-chains)))))

;; ---
;; --- Test definitions
;; ---
(defn test-keypers-dkg-generation
  [{:keys [num-keypers] :as conf}]
  {:test/id :keyper-dkg-works
   :test/conf conf
   :test/description "distributed key generation should work"
   :test/steps [{:run :init/init
                 :init/conf conf}
                (for [keyper (range num-keypers)]
                  {:check :keyper/meta-inf
                   :keyper-num keyper})

                (build/run-chain)
                (build/run-node conf)
                (build/run-keypers conf)

                {:run :process/run
                 :process/id :boot
                 :process/cmd '[bb boot]
                 :process/wait true}

                {:check :loop/until
                 :loop/description "eon should exist for all keypers"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range num-keypers)]
                                {:check :keyper/eon-exists
                                 :keyper/num keyper
                                 :keyper/eon 1})}

                {:check :loop/until
                 :loop/description "all keypers should succeed with the dkg process"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range num-keypers)]
                                {:check :keyper/dkg-success
                                 :keyper/eon 1
                                 :keyper/num keyper})}

                (for [keyper (range num-keypers)]
                  {:check :keyper/non-zero-activation-block
                   :keyper/num keyper})

                ]})

(defn test-change-keyper-set
  []
  (let [num-keypers 4
        num-initial-keypers (dec num-keypers)
        devmode? false
        conf {:num-keypers num-keypers
              :num-initial-keypers num-initial-keypers}]
    {:test/id :change-keyper-set
     :test/conf conf
     :test/description "changing the keyper set should work"
     :test/steps [{:run :init/init
                   :init/conf conf}
                  (for [keyper (range num-keypers)]
                    {:check :keyper/meta-inf
                     :keyper-num keyper})

                  (if devmode?
                    (build/run-chain)
                    {:run :chain/run-chains})

                  [{:run :process/run
                    :process/id :node
                    :process/cmd '[bb node]
                    :process/opts {:extra-env {"PLAY_NUM_KEYPERS" num-initial-keypers}}
                    :process/port 8545
                    :process/port-timeout (+ 5000 (* num-keypers 2000))}

                   {:run :process/run
                    :process/wait true
                    :process/id :symlink-deployments
                    :process/cmd '[bb -deployments]}]

                  (build/run-keypers conf)
                  (when devmode?
                    {:run :process/run
                     :process/id :boot
                     :process/cmd '[bb boot]
                     :process/wait true})

                  {:check :loop/until
                   :loop/description "All keypers should see the new keyper_set"
                   :loop/timeout-ms (* 20 1000)
                   :loop/checks (for [keyper (range num-keypers)]
                                  {:check :keyper/keyper-set
                                   :keyper/num keyper
                                   :keyper/expected-count 2})}

                  {:check :loop/until
                   :loop/description "eon should exist for all keypers"
                   :loop/timeout-ms (* 60 1000)
                   :loop/checks (for [keyper (range num-initial-keypers)]
                                  {:check :keyper/eon-exists
                                   :keyper/num keyper
                                   :keyper/eon 1})}

                  {:check :loop/until
                   :loop/description "all keypers should succeed with the dkg process"
                   :loop/timeout-ms (* 60 1000)
                   :loop/checks (for [keyper (range num-initial-keypers)]
                                  {:check :keyper/dkg-success
                                   :keyper/eon 1
                                   :keyper/num keyper})}

                  (for [keyper (range num-initial-keypers)]
                    {:check :keyper/non-zero-activation-block
                     :keyper/num keyper})

                  {:run ::add-spare-keyper-set}
                  {:run ::configure-keypers}

                  {:check :loop/until
                   :loop/description "All keypers should notice the configuration change"
                   :loop/timeout-ms (* 20 1000)
                   :loop/checks (for [keyper (range num-keypers)]
                                  {:check :keyper/keyper-set
                                   :keyper/num keyper
                                   :keyper/expected-count 3})}

                  {:check :loop/until
                   :loop/description "eon 2 should exist for all keypers"
                   :loop/timeout-ms (* 60 1000)
                   :loop/checks (for [keyper (range num-keypers)]
                                  {:check :keyper/eon-exists
                                   :keyper/num keyper
                                   :keyper/eon 2})}

                  {:check :loop/until
                   :loop/description "all keypers should succeed with the dkg process"
                   :loop/timeout-ms (* 60 1000)
                   :loop/checks (for [keyper (range num-keypers)]
                                  {:check :keyper/dkg-success
                                   :keyper/eon 2
                                   :keyper/num keyper})}
                  (for [keyper (range num-keypers)]
                    {:check :keyper/tendermint-batch-config-started
                     :keyper-num keyper})
                  ]}))

(defn test-dkg-keypers-join-late
  [{:keys [num-keypers threshold] :as conf}]
  {:test/id :late-keyper-dkg-works
   :test/conf conf
   :test/description "distributed key generation should work when a keyper joins late"
   :test/steps [{:run :init/init
                 :init/conf conf}
                (for [keyper (range num-keypers)]
                  {:check :keyper/meta-inf
                   :keyper-num keyper})
                (build/run-chain)
                (build/run-node conf)

                (mapv build/run-keyper (range (dec threshold)))

                {:run :process/run
                 :process/id :boot
                 :process/cmd '[bb boot]
                 :process/wait true}

                {:check :loop/until
                 :loop/description "eon should exist for all keypers"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range (dec threshold))]
                                {:check :keyper/eon-exists
                                 :keyper/num keyper
                                 :keyper/eon 1})}

                {:check :loop/until
                 :loop/description "all keypers should fail the dkg process"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range (dec threshold))]
                                {:check :keyper/dkg-failed
                                 :keyper/num keyper})}

                ;; start the late keypers
                (mapv build/run-keyper (range (dec threshold) num-keypers))
                {:check :loop/until
                 :loop/description "all keypers should see the new eon"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range num-keypers)]
                                {:check :keyper/eon-exists
                                 :keyper/num keyper
                                 :keyper/eon 2})}

                {:check :loop/until
                 :loop/description "all keypers should succeed with the dkg process"
                 :loop/timeout-ms (* 60 1000)
                 :loop/checks (for [keyper (range num-keypers)]
                                {:check :keyper/dkg-success
                                 :keyper/eon 2
                                 :keyper/num keyper})}

                (for [keyper (range num-keypers)]
                  {:check :keyper/non-zero-activation-block
                   :keyper/num keyper})

                ]})

(defn generate-tests
  []
  (concat
   [(test-change-keyper-set)]
   (for [conf [{:num-keypers 3, :threshold 2}]
         f [test-keypers-dkg-generation
            test-dkg-keypers-join-late]]
     (f conf))))

(def tests (delay (generate-tests)))
