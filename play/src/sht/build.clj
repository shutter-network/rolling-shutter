(ns sht.build
  "sht.build contains some functions that build steps for use in tests"
  (:require [sht.play :as play]))

(defn init
  [{:keys [num-keypers num-decryptors]}]
  [{:run :process/run
    :process/id :build
    :process/cmd '[bb build]
    :process/wait true}
   {:run :process/run
    :process/id :init
    :process/cmd '[bb init]
    :process/wait true}])

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
