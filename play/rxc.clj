(ns rxc
  (:require [clojure.core.server :as server]
            [clojure.java.io     :as io]
            [babashka.process :as p]
            [babashka.fs :as fs])
  (:import [java.net ConnectException Socket]))

(def *process-map* (atom {}))

(defn start-prepl-server!
  [port]
  (let [host "0.0.0.0"
        socket         (server/start-server
                        {:accept `server/io-prepl
                         :address host
                         :port    port
                         :name    "Dave's amazing prepl server"})
        effective-port (.getLocalPort socket)]

    (println "Started prepl server on port" effective-port)

    ;; Wait until process is interrupted.
    @(promise)))

(defn list-processes
  []
  (keys @*process-map*))

(defn wait-proc
  [proc-id]
  (if-let [proc (get @*process-map* proc-id)]
    @proc
    nil))

(defn start-process
  [proc-id cmd opts]
  (let [proc (p/process cmd opts)]
    (swap! *process-map* assoc proc-id proc)
    nil))

(start-prepl-server! 1666)
