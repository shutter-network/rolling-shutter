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

(defn sys-write-config-files
  "write config files of all subprocesses to disk.
  This needs to be called, when :toml-edits has been modified."
  [{:sys/keys [keypers collator mocksequencer p2pnodes] :as sys}]
  (doseq [sub (concat keypers [collator mocksequencer] p2pnodes)]
    (play/toml-edit-file (:subcommand/cfgfile sub)
                         (:subcommand/toml-edits sub))))

(defmethod runner/run :init/init
  [sys {:init/keys [conf] :as m}]
  (info "Initializing system" m)
  (let [sys (bb-build-all sys)
        {:keys [num-keypers num-bootstrappers]} conf
        keypers (->> (range num-keypers)
                     (map play/keyper-subcommand)
                     (map play/generate-config)
                     (mapv play/initdb))
        collator (-> (play/collator-subcommand)
                     play/generate-config
                     play/initdb)
        p2pnodes (->> (range num-bootstrappers)
                      (map play/p2pnode-subcommand)
                      (mapv play/generate-config))

        mocksequencer (play/generate-config (play/mocksequencer-subcommand))

        get-bootstrap-addrs (fn [p2pnode]
                              (play/construct-boostrap-addresses
                               (merge
                                (get-in p2pnode [:subcommand/cfg])
                                {:listen-addrs (get-in p2pnode [:subcommand/toml-edits "ListenAddresses"])})))
        set-peers (fn [sub]
                    (assoc-in sub  [:subcommand/toml-edits "CustomBootstrapAddresses"]
                              (flatten (map get-bootstrap-addrs p2pnodes))))
        keypers (map set-peers keypers)
        collator (set-peers collator)
        p2pnodes (map set-peers p2pnodes)
        res {:sys/keypers keypers
             :sys/collator collator
             :sys/mocksequencer mocksequencer
             :sys/p2pnodes p2pnodes}
        sys (merge sys res)]
    (println p2pnodes)
    (sys-write-config-files sys)
    (info "Initialized system successfully" res)
    sys))

(defn sys-deploy-conf
  "create a deploy conf structure as read by the 'hardhat node' command"
  [sys]
  (let [eth-address (fn [m] (get-in m [:subcommand/cfg :eth-address]))]
    {:keypers (mapv eth-address (:sys/keypers sys))
     :collator (eth-address (:sys/collator sys))
     :fundValue "100"}))

(defn run-node
  [{:keys [num-keypers]}]
  [{:run :process/run
    :process/id :node
    :process/cmd '[bb node]
    :process/port 8545
    :process/port-timeout (+ 15000 (* num-keypers 2000))}

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

(defn run-p2pnode
  [n]
  (let [p2pnode (play/p2pnode-subcommand n)]
    {:run :process/run
     :process/id (keyword (format "p2pnode-%d" n))
     :process/cmd (play/subcommand-run p2pnode)
     :process/port (:subcommand/p2p-port p2pnode)
     :process/port-timeout 3000}))

(defn run-p2pnodes
  [{:keys [num-bootstrappers]}]
  (mapv run-p2pnode (range num-bootstrappers)))

(defn run-collator
  []
  (let [collator (play/collator-subcommand)]
    {:run :process/run
     :process/id :collator
     :process/cmd (play/subcommand-run collator)
     :process/port (:subcommand/p2p-port collator)
     :process/port-timeout 3000}))

(defn run-mocksequencer
  []
  (let [mock-sequencer (play/mocksequencer-subcommand)]
    {:run :process/run
     :process/id :mocksequencer
     :process/cmd (play/subcommand-run mock-sequencer)
     :process/port (:subcommand/listening-port mock-sequencer)
     :process/port-timeout 3000}))

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
