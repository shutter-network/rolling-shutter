;; this namespace is read by babashka via our bb.edn file. Make sure it stays compatible with
;; babashka
(ns sht.play
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]
            [clojure.java.shell]
            [clojure.pprint :as pprint]
            [clojure.string :as str]
            [cheshire.core :as json]
            [babashka.process :as p]
            [babashka.fs :as fs]))

(def ^:private base-port 23000)
(def ^:private keyper-base-port (+ base-port 100))
(def ^:private ethereum-rpc-port 8545)
;; use the "layer 1" ethereum node for the contracts
(def ^:private contracts-rpc-port ethereum-rpc-port)
(def ^:private sequencer-rpc-port 8555)

(def ^:dynamic *cwd* (str (fs/normalize (fs/absolutize "."))))

(def repo-root
  (str (fs/canonicalize (or (System/getenv "ROLLING_SHUTTER_ROOT") ".."))))

(defn- split-path []
  (str/split (System/getenv "PATH") (re-pattern java.io.File/pathSeparator)))

(defn- join-path
  [ps]
  (str/join java.io.File/pathSeparator (distinct ps)))

(defn shutter-env []
  {"ROLLING_SHUTTER_ROOT" repo-root
   "PATH" (->> (split-path)
               (cons (str (fs/path repo-root "rolling-shutter" "bin")))
               join-path)})

(def rolling-shutter (str (fs/path repo-root "rolling-shutter" "bin" "rolling-shutter")))
;; (def rolling-shutter "rolling-shutter")

(alter-var-root
 #'p/*defaults*
 (fn [m]
   (assoc m :env (merge {} (System/getenv) (shutter-env)))))

(defn bb-log
  [& args])

(defn set-bb-log!
  [log]
  (alter-var-root #'bb-log (fn [_] log)))

(defn replace-rolling-shutter-absolute-path
  [cmd]
  (if (= (first cmd) 'rolling-shutter)
    (cons rolling-shutter (rest cmd))
    cmd))

(defn process
  ([cmd]
   (process cmd {}))
  ([cmd opts]
   (let [opts (merge {:dir *cwd*} opts)]
     (bb-log (seq cmd) opts)
     (p/process (replace-rolling-shutter-absolute-path cmd) opts))))

(defn run-process
  ([cmd]
   (run-process cmd {}))
  ([cmd {:keys [dir] :as opts}]
   (when dir
     (println (format "Entering directory '%s'" dir)))
   (try
     (bb-log (seq cmd) opts)
     (p/check (p/process (replace-rolling-shutter-absolute-path cmd)
                         (merge {:out :inherit :err :inherit}
                                opts)))
     (finally
       (when dir
         (println (format "Leaving directory '%s'" dir)))))))

(def dropdb-with-force?
  (delay
    (str/includes? (:out (clojure.java.shell/sh "dropdb" "--help"))
                   "--force")))

(defn dropdb
  [db]
  (p/check @(process (concat ["dropdb" "--if-exists"]
                             (if @dropdb-with-force?
                               ["--force" db]
                               [db])))))

(defn- extract-address [toml]
  (re-find (re-pattern "0x[0-9a-fA-F]{40}") toml))

(defn- extract-peerid [toml]
  (second (re-find (re-pattern "(?m)(?i)^# Peer identity: /p2p/([0-9a-zA-Z]*)") toml)))

(defn- extract-eon-key [toml]
  (second (re-find (re-pattern "(?m)(?i)^# Eon Public Key: ([0-9a-zA-Z]*)") toml)))

(defn- extract-cfg [cfgfile]
  (let [toml (slurp (str (fs/path *cwd* cfgfile)))]
    {:eth-address (extract-address toml)
     :peerid (extract-peerid toml)
     :eon-key (extract-eon-key toml)}))

(defn toml-replace
  [toml-str key value]
  (str/replace-first
   toml-str
   (re-pattern (format "(?m)(^\\s*%s\\s*=)(.*)" key))
   (str "$1 " (json/encode value))))

(defn toml-edit-string
  [toml-str m]
  (reduce (fn [toml-str [k v]]
            (toml-replace toml-str k v))
          toml-str
          m))

(defn toml-edit-file
  [filename m]
  (let [filename (str (fs/path *cwd* filename))]
    (spit filename (toml-edit-string (slurp filename) m))))

(defn subcommand-run
  [{:subcommand/keys [cmd cfgfile]}]
  ['rolling-shutter (str cmd) "--config" cfgfile])

(defn subcommand-genconfig
  [{:subcommand/keys [cmd cfgfile]}]
  ['rolling-shutter (str cmd) "generate-config" "--output" cfgfile])

(defn generate-config
  [{:subcommand/keys [toml-edits cfgfile] :as subcommand}]
  (when-not (fs/exists? (fs/path *cwd* cfgfile))
    (let [cmd (subcommand-genconfig subcommand)]
      (p/check @(process cmd {:out :string :err :string}))
      (toml-edit-file cfgfile toml-edits)))
  (assoc subcommand
         ::cwd *cwd*
         :subcommand/cfg (extract-cfg cfgfile)))

(defn initdb
  [{:subcommand/keys [cmd cfgfile db] :as subcommand}]
  (dropdb db)
  (p/check (process ["createdb" db]))
  (p/check (process ['rolling-shutter cmd "initdb" "--config" cfgfile]))
  subcommand)

;; -- keyper-subcommand
(defn keyper-subcommand
  [n]
  (let [db (format "keyper-db-%d" n)
        p2p-port (+ keyper-base-port n)]
    #:subcommand{:cmd 'keyper
                 :db db
                 :p2p-port p2p-port
                 :cfgfile (format "keyper-%s.toml" n)
                 :toml-edits {"DatabaseURL" (format "postgres:///%s" db)
                              "DKGPhaseLength" 8
                              "ListenAddress" (format "/ip4/127.0.0.1/tcp/%d" p2p-port)
                              "HTTPEnabled" true
                              "HTTPListenAddress" (format ":%d" (+ 24000 n))
                              "ContractsURL" (format "http://127.0.0.1:%d/" ethereum-rpc-port)}}))

;; -- mocknode-subcommand
(defn mocknode-subcommand
  []
  (let [p2p-port (+ base-port 0)]
    #:subcommand{:cmd 'mocknode
                 :cfgfile "mock.toml"
                 :p2p-port p2p-port
                 :db nil
                 :toml-edits {"ListenAddress" (format "/ip4/127.0.0.1/tcp/%d" p2p-port)}}))

;; -- collator
(defn collator-subcommand
  []
  (let [p2p-port (+ base-port 1)]
    #:subcommand{:cmd 'collator
                 :cfgfile "collator.toml"
                 :p2p-port p2p-port
                 :db "collator"
                 :toml-edits {"DatabaseURL" (format "postgres:///collator")
                              "ListenAddress" (format "/ip4/127.0.0.1/tcp/%d" p2p-port)
                              "SequencerURL" (format "http://localhost:%d" sequencer-rpc-port)}}))

;; -- mocksequencer
(defn mocksequencer-subcommand
  []
  #:subcommand{:cmd 'mock-sequencer
               :cfgfile "mocksequencer.toml"
               :listening-port sequencer-rpc-port
               :toml-edits {"EthereumURL" (format "http://localhost:%d" ethereum-rpc-port)
                            "ContractsURL" (format "http://localhost:%d" contracts-rpc-port)
                            "HTTPListenAddress" (format ":%d" sequencer-rpc-port)}})

(defn ci-gen
  "Rewrite bb.edn with a simplified build for use on CI systems"
  []
  (let [src "bb.edn"
        dst "ci-bb.edn"
        bb (-> (edn/read-string (slurp src))
               (assoc-in [:tasks 'build] :do-nothing)
               (assoc-in [:tasks 'contracts:install] :do-nothing))]
    (with-open [w (io/writer dst)]
      (pprint/pprint bb w))
    (println "Created simpified config in" dst)))

(def keyper-db (comp :subcommand/db keyper-subcommand))
(def collator-db (comp :subcommand/db collator-subcommand))
