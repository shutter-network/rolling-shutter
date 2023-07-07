;; this namespace is read by babashka via our bb.edn file. Make sure it stays compatible with
;; babashka
(ns sht.play
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]
            [clojure.java.shell]
            [clojure.pprint :as pprint]
            [clojure.string :as str]
            [cheshire.core :as json]
            [babashka.http-client :as http]
            [babashka.process :as p]
            [babashka.fs :as fs]))

(def ^:private base-port 23000)
(def ^:private keyper-base-port (+ base-port 100))
(def ^:private bootstrap-base-port (+ base-port 200))
(def ^:private ethereum-rpc-port 8545)
;; use the "layer 1" ethereum node for the contracts
(def ^:private contracts-rpc-port ethereum-rpc-port)
(def ^:private sequencer-rpc-port 8555)
(def ^:private default-loglevel "debug")

(def ^:dynamic *cwd* (str (fs/normalize (fs/absolutize "."))))

(def repo-root
  (let [candidates [(System/getenv "ROLLING_SHUTTER_ROOT")
                    (fs/path (System/getProperty "babashka.config") ".." "..")
                    ".."]
        candidates (->> candidates
                        (remove nil?)
                        (map fs/canonicalize)
                        (map str)
                        distinct)
        root? (fn [p]
                (fs/exists? (fs/path p "rolling-shutter" "keyper" "keyper.go")))
        root (first (filter root? candidates))]
    (if root
      root
      (throw (ex-info "could not determine root directory" {:candidates candidates})))))

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

(defn- run-process*
  [cmd {:keys [dir] :as opts}]
  (when dir
    (println (format "Entering directory '%s'" dir)))
  (try
    (bb-log (seq cmd) opts)
    (deref (p/process (replace-rolling-shutter-absolute-path cmd)
                      (merge {:out :inherit :err :inherit}
                             opts)))
    (finally
      (when dir
        (println (format "Leaving directory '%s'" dir))))))

(defn run-process
  ([cmd]
   (run-process cmd {}))
  ([cmd {:keys [dir] :as opts}]
   (let [proc (run-process* cmd opts)
         exit-code (:exit proc)]
     (when-not (zero? exit-code)
       (println (format "Error: %s returned with non-zero exit code %d" (first cmd) exit-code))
       (System/exit 1)))))

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

(defn extract-address [toml]
  (re-find (re-pattern "0x[0-9a-fA-F]{40}") toml))

(defn extract-peerid [toml]
  (second (re-find (re-pattern "(?m)(?i)^# Peer identity: /p2p/([0-9a-zA-Z]*)") toml)))

(defn extract-peer-role [toml]
  (second (re-find (re-pattern "(?m)(?i)^# Peer role: ([a-zA-Z]*)") toml)))

(defn extract-eon-key [toml]
  (second (re-find (re-pattern "(?m)(?i)^# Eon Public Key: ([0-9a-zA-Z]*)") toml)))

(defn extract-toml-value [toml key]
  (second (re-find (re-pattern (format "(?m)^\\s*%s\\s*=\\s*(.*)" key)) toml)))

(defn remove-outer-characters [string]
  (subs string 1 (dec (count string))))

(defn extract-listen-addresses [toml]
  (json/decode (extract-toml-value toml "ListenAddresses")))

(defn extract-cfg [path]
  (let [toml (slurp (str path))]
    {:eth-address (extract-address toml)
     :peerid (extract-peerid toml)
     :listen-addrs (extract-listen-addresses toml)
     :peer-role (extract-peer-role toml)
     :eon-key (extract-eon-key toml)}))

(defn construct-boostrap-addresses [cfg]
  (map (fn [listen-addr]
         (str listen-addr
              "/p2p/"
              (get cfg :peerid)))
       (get cfg :listen-addrs)))

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
  (let [path (str (fs/path *cwd* filename))]
    (spit path (toml-edit-string (slurp path) m))))

(defn subcommand-run
  [{:subcommand/keys [cmd cfgfile loglevel]}]
  ['rolling-shutter
   (str "--loglevel")
   (str (or loglevel default-loglevel))
   (str cmd)
   "--config" cfgfile])

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
         :subcommand/cfg (extract-cfg (fs/path *cwd* cfgfile))))

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
                 :loglevel nil
                 :db db
                 :p2p-port p2p-port
                 :cfgfile (format "keyper-%s.toml" n)
                 :toml-edits {"DatabaseURL" (format "postgres:///%s" db)
                              "DKGPhaseLength" 8
                              "DKGStartBlockDelta" 5
                              "ListenAddresses" [(format "/ip4/127.0.0.1/tcp/%d" p2p-port)]
                              "HTTPEnabled" true
                              "Environment" "local"
                              "HTTPListenAddress" (format ":%d" (+ 24000 n))
                              "EthereumURL" (format "http://127.0.0.1:%d/" ethereum-rpc-port)
                              "ContractsURL" (format "http://127.0.0.1:%d/" contracts-rpc-port)}}))

;; -- mocknode-subcommand
(defn mocknode-subcommand
  []
  (let [p2p-port (+ base-port 0)]
    #:subcommand{:cmd 'mocknode
                 :loglevel nil
                 :cfgfile "mock.toml"
                 :p2p-port p2p-port
                 :db nil
                 :toml-edits {"ListenAddresses" [(format "/ip4/127.0.0.1/tcp/%d" p2p-port)]}}))

;; -- p2pnode
(defn p2pnode-subcommand
  [n]
  (let [p2p-port (+ bootstrap-base-port n)]
    #:subcommand{:cmd 'p2pnode
                 :loglvl nil
                 :cfgfile (format "p2p-%s.toml" n)
                 :p2p-port p2p-port
                 :db nil
                 :toml-edits {"ListenAddresses" [(format "/ip4/127.0.0.1/tcp/%d" p2p-port)]
                              "Environment" "local"}}))

;; -- collator
(defn collator-subcommand
  []
  (let [p2p-port (+ base-port 1)]
    #:subcommand{:cmd 'collator
                 :loglevel nil
                 :cfgfile "collator.toml"
                 :p2p-port p2p-port
                 :db "collator"
                 :toml-edits {"DatabaseURL" (format "postgres:///collator")
                              "ListenAddresses" [(format "/ip4/127.0.0.1/tcp/%d" p2p-port)]
                              "Environment" "local"
                              "EthereumURL" (format "http://127.0.0.1:%d/" ethereum-rpc-port)
                              "ContractsURL" (format "http://127.0.0.1:%d/" contracts-rpc-port)
                              "SequencerURL" (format "http://localhost:%d" sequencer-rpc-port)}}))

;; -- mocksequencer
(defn mocksequencer-subcommand
  []
  #:subcommand{:cmd 'mocksequencer
               :loglevel nil
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

(defn jsonrpc-body
  [method params]
  (json/generate-string {:jsonrpc "2.0"
                         :method method
                         :params params
                         :id 1}))

(defn get-jsonrpc
  [url]
  (http/get url)
)

(defn post-jsonrpc
  [url body]
  (http/post url
             {:headers {"Content-Type" "application/json"}
              :body body}))

(defn get-jsonrpc-result
  [resp]
  (let [{:keys [result error]} (json/parse-string (:body resp) true)]
    (when error
      (throw (ex-info (:message error) {:code (:code error)})))
    result))

(defn add-collator
  [mocksequencer-url addr l1-blocknumber]
  (->> (jsonrpc-body "admin_addCollator" [addr l1-blocknumber])
       (post-jsonrpc mocksequencer-url)
       get-jsonrpc-result))

(comment
  (post-jsonrpc "http://localhost:9999"
                (jsonrpc-body "admin_addCollator"
                              ["0x96858D19fB1398a23fd3c5E9fb205B964d5BA46b" 55]))
  (add-collator "http://localhost:8555" "0x96858D19fB1398a23fd3c5E9fb205B964d5BA46b" 55))
