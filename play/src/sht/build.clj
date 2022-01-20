(ns sht.build
  "sht.build contains some functions that build steps for use in tests"
  (:require [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [sht.play :as play]
            [sht.runner :as runner]))

(defn- bb-build-all
  "bb-build-all makes sure we use an up-to-date executable by calling the relevant babashka
  task. This is one of the steps, where it's fine to use babashka."
  [sys]
  (runner/dispatch sys
                   [{:run :process/run
                     :process/id :bb-build-all
                     :process/cmd '[bb build-all]
                     :process/wait true}]))

(defmethod runner/run :init/init
  [sys {:init/keys [conf] :as m}]
  (info "Initializing system" m)
  (let [sys (bb-build-all sys)
        {:keys [num-keypers num-decryptors]} conf
        keypers (->> (range num-keypers)
                     (map play/keyper-subcommand)
                     (map play/generate-config)
                     (mapv play/initdb))
        decryptors (->> (range num-decryptors)
                        (map (partial play/decryptor-subcommand num-decryptors))
                        (map play/generate-config)
                        (mapv play/initdb))
        collator (-> (play/collator-subcommand)
                     play/generate-config
                     play/initdb)
        res {:sys/keypers keypers
             :sys/decryptors decryptors
             :sys/collator collator}]
    (info "Initialized system successfully" res)
    (merge sys res)))

(defn run-node
  [{:keys [num-keypers]}]
  [{:run :process/run
    :process/id :node
    :process/cmd '[bb node]
    :process/port 8545
    :process/port-timeout (+ 5000 (* num-keypers 2000))}

   {:run :process/run
    :process/wait true
    :process/id :symlink-deployments
    :process/cmd '[bb -deployments]}])

(defn run-keyper
  [n]
  (let [keyper (play/keyper-subcommand n)]
    {:run :process/run
     :process/id (keyword (format "keyper-%d" n))
     :process/cmd (play/subcommand-run keyper)
     :process/port (:subcommand/p2p-port keyper)
     :process/port-timeout 3000}))

(defn run-keypers
  [{:keys [num-keypers]}]
  (mapv run-keyper (range num-keypers)))

(defn run-chain
  []
  [{:run :process/run
    :process/wait true
    :process/id :init-chain
    :process/cmd ['rolling-shutter "chain" "init" "--root" "testchain" "--dev" "--blocktime" "1"]}
   {:run :process/run
    :process/id :chain
    :process/cmd ['rolling-shutter "chain" "--config" "testchain/config/config.toml"]
    :process/port 26657}])
