(ns sht.dkg-test
  (:require [next.jdbc :as jdbc]
            [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [cheshire.core :as json]
            [sht.runner :as runner]
            [sht.build :as build]
            [babashka.fs :as fs]
            [sht.play :as play]))

(defonce play-db-password (or (System/getenv "PLAY_DB_PASSWORD") ""))

(defmethod runner/check :keyper/meta-inf
  [sys {:keys [keyper-num]}]
  (let [db {:dbtype "postgresql"
            :dbname (play/keyper-db keyper-num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        row (jdbc/execute-one! ds ["select * from meta_inf"])]
    {:chk/ok? (some? row)
     :chk/description "meta_inf table should be filled"
     :chk/info row}))

(defmethod runner/check :keyper/eon-exists
  [sys {:keyper/keys [num eon]}]
  (let [db {:dbtype "postgresql"
            :dbname (play/keyper-db num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        row (jdbc/execute-one! ds ["select * from eons where eon=?" eon])]
    {:chk/ok? (some? row)
     :chk/description "eon should exist"
     :chk/info row
     :keyper/num num}))

(defmethod runner/check :keyper/dkg-success
  [sys {:keyper/keys [num]}]
  (let [db {:dbtype "postgresql"
            :dbname (play/keyper-db num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        rows (jdbc/execute-one! ds ["select * from dkg_result where success"])]
    {:chk/ok? (not (empty? rows))
     :chk/description "dkg should finish successfully"
     :chk/info rows
     :keyper/num num}))

(defmethod runner/check :keyper/dkg-failed
  [sys {:keyper/keys [num]}]
  (let [db {:dbtype "postgresql"
            :dbname (play/keyper-db num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        rows (jdbc/execute-one! ds ["select * from dkg_result where not success"])]
    {:chk/ok? (not (empty? rows))
     :chk/description "dkg should fail"
     :chk/info rows
     :keyper/num num}))

(defmethod runner/check :keyper/non-zero-activation-block
  [sys {:keyper/keys [num]}]
  (let [db {:dbtype "postgresql"
            :dbname (play/keyper-db num)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        rows (jdbc/execute-one! ds ["select * from eons where activation_block_number<=0"])]
    {:chk/ok? (empty? rows)
     :chk/description "activation block number must be positive"
     :chk/info rows
     :keyper/num num}))

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
                                 :keyper/num keyper})}

                (for [keyper (range num-keypers)]
                  {:check :keyper/non-zero-activation-block
                   :keyper/num keyper})

                ]})


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

(defn test-change-keyper-set
  []
  (let [num-keypers 4
        num-initial-keypers (dec num-keypers)
        conf {:num-keypers num-keypers
              :num-decryptors 0}]
    {:test/id :change-keyper-set
     :test/conf conf
     :test/description "distributed key generation should work"
     :test/steps [{:run :init/init
                   :init/conf conf}
                  (for [keyper (range num-keypers)]
                    {:check :keyper/meta-inf
                     :keyper-num keyper})

                  (build/run-chain)
                  [{:run :process/run
                    :process/id :node
                    :process/cmd '[bb node]
                    :process/opts {:extra-env {"PLAY_NUM_KEYPERS" num-initial-keypers
                                               "PLAY_NUM_DECRYPTORS" "0"}}
                    :process/port 8545
                    :process/port-timeout (+ 5000 (* num-keypers 2000))}

                   {:run :process/run
                    :process/wait true
                    :process/id :symlink-deployments
                    :process/cmd '[bb -deployments]}]
                  ;; (build/run-node conf)

                  (build/run-keypers conf)

                  {:run :process/run
                   :process/id :boot
                   :process/cmd '[bb boot]
                   :process/wait true}

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
                                   :keyper/num keyper})}

                  (for [keyper (range num-initial-keypers)]
                    {:check :keyper/non-zero-activation-block
                     :keyper/num keyper})

                  {:run ::configure-keypers}

                  {:check :loop/until
                   :loop/description "eon 2 should exist for all keypers"
                   :loop/timeout-ms (* 60 1000)
                   :loop/checks (for [keyper (range num-keypers)]
                                  {:check :keyper/eon-exists
                                   :keyper/num keyper
                                   :keyper/eon 2})}

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
                                 :keyper/num keyper})}

                (for [keyper (range num-keypers)]
                  {:check :keyper/non-zero-activation-block
                   :keyper/num keyper})

                ]})

(defn generate-tests
  []
  (for [conf [{:num-keypers 3, :num-decryptors 2, :threshold 2}]
        f [test-keypers-dkg-generation
           test-dkg-keypers-join-late]]
    (f conf)))

(def tests (delay (generate-tests)))
