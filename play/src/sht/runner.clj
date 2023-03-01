(ns sht.runner
  "sht.runner provides the means to run sytem tests. Each test is just a simple map, with the
  following fields:

  - test/id:            a keyword identifying the test
  - test/description:   a textual description of what we test
  - test/steps:         a sequence of steps to execute

  Each step can be a list of steps, which are are then recursively executed, or a map containing
  either an `:run` step or an `:check` step. The former describes an action to be executed, which
  changes the state of the running system. The latter describes an observation to be made about the
  system, these should not change the state of the system. Normally we observe the state by
  executing queries against the different postgresql databases.

  `:run` steps
  A `:run` step is a map with a `:run` key. The test runner will call the `sht.runner/run`
  multimethod, which dispatches via the value of the `:run` key. One common `:run` step is to start
  external processes. An example map looks like:

    {:run :process/run
     :process/id :node
     :process/cmd '[bb node]
     :process/port 8545
     :process/port-timeout (+ 5000 (* num-keypers 2000))}

  This starts 'bb node', and waits with the given timeout for the node process to start listening
  on the given port.

  `:check` steps
  A `:check` step is a map with a `:check` key. The test runner will call the `sht.runner/check`
  multimethod, which dispatches via the value of the `:check` key. The method must return a map
  with the following keys describing the observation:

  - chk/ok?:           a boolean, whether the check has been successful
  - chk/description:   a textual description of what we checked
  - chk/info:          arbitrary data about the check

  One common `:check` step is to wait for a sequence of `:check` substeps with a timeout:

    {:check :loop/until
     :loop/description \"eon should exist for all keypers\"
     :loop/timeout-ms (* 60 1000)
     :loop/checks (for [keyper (range (dec threshold))]
                     {:check :keyper/eon-exists
                      :keyper/num keyper
                      :keyper/eon 1})}

  Using multimethods gives us the ability to extend the set of possible `:run` and `:check` steps
  easily.
"
  (:require [clojure.pprint :as pprint]
            [clojure.string :as str]
            [clojure.java.io :as io]
            [puget.printer :as puget]
            [babashka.process :as p]
            [babashka.fs :as fs]
            [sht.play :as play]
            [taoensso.encore :as enc]
            [taoensso.timbre.appenders.core]
            [taoensso.timbre :as timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]))

;; Allow pretty printing of process objects, see https://github.com/babashka/process#clojurepprint
(prefer-method pprint/simple-dispatch clojure.lang.IPersistentMap clojure.lang.IDeref)

(do
  ;; SQL arrays as clojure vectors, see
  ;; https://cljdoc.org/d/seancorfield/next.jdbc/1.2.659/doc/getting-started/tips-tricks#working-with-arrays
  (import  '[java.sql Array])
  (require '[next.jdbc.result-set :as rs])

  (extend-protocol rs/ReadableColumn
    Array
    (read-column-by-label [^Array v _]    (vec (.getArray v)))
    (read-column-by-index [^Array v _ _]  (vec (.getArray v)))))

(def empty-report
  #:report{:num-actions 0
           :num-checks-failed 0
           :num-checks-succeeded 0
           :checks []})

(defn report-add-check-result
  [report check]
  (-> report
      (update :report/checks conj check)
      (update (if (:chk/ok? check)
                :report/num-checks-succeeded
                :report/num-checks-failed)
              inc)))

(defn report-count-action
  [report]
  (update report :report/num-actions inc))

(defn timeout
  "Run the given `callback` function in a future and return the result or cancel the future after
  `timeout-ms` milliseconds."
  [timeout-ms callback]
  (let [fut (future (callback))
        ret (deref fut timeout-ms ::timed-out)]
    (when (= ret ::timed-out)
      (future-cancel fut))
    ret))

(def ^:dynamic *process-map* nil)

(defn start-proc!
  [sys proc-id cmd opts]
  (info "starting process" proc-id cmd)
  (let [log-dir (:log-dir sys)
        opts (merge {:extra-env {"PLAY_NUM_KEYPERS" (-> sys :conf :num-keypers str)}
                     :out (-> log-dir (fs/path (format "%s-out.txt" (name proc-id))) str io/file),
                     :err (-> log-dir (fs/path (format "%s-err.txt" (name proc-id))) str io/file)}
                    {:dir (:cwd sys)}
                    opts)
        cmd (play/replace-rolling-shutter-absolute-path cmd)
        cmd (if (= 'bb (first cmd))
              (concat ["bb" "--config" (:bb-edn sys)]
                      (rest cmd))
              cmd)
        proc (p/process cmd opts)]
    (swap! *process-map* assoc proc-id proc)
    sys))

(defn wait-proc!
  [sys proc-id]
  (info "waiting for process" proc-id)
  (let [proc (deref (get @*process-map* proc-id))]
    (swap! *process-map* assoc proc-id proc)
    (p/check proc)
    sys))

(defn wait-file-forever
  [p]
  (when-not (fs/exists? p)
    (Thread/sleep 50)
    (recur p)))

(defn wait-file
  [p timeout-ms]
  (when (= ::timed-out (timeout timeout-ms (partial wait-file-forever p)))
    (throw (ex-info (format "Timeout waiting for file %s" p)
                    {:path p
                     :timeout-ms timeout-ms}))))

(defn- wait-port-forever
  [host port]
  (when (try
          (with-open [_ (java.net.Socket. host port)]
            nil)
          (catch java.net.ConnectException err
            err))
    (Thread/sleep 50)
    (recur host port)))

(defn wait-port
  ([sys port]
   (wait-port sys port {}))
  ([sys port {:keys [host timeout-ms] :or {host "127.0.0.1" timeout-ms 5000}}]
   (when (= ::timed-out (timeout timeout-ms (partial wait-port-forever host port)))
     (throw (ex-info (format "Timeout waiting for port %d" port)
                     {:port port
                      :host host
                      :timeout-ms timeout-ms})))
   sys))

(defmulti run
  "Run the given step"
  (fn [sys m] (:run m)))
(defmethod run :default run-default
  [sys m]
  (throw (ex-info "cannot run" {:m m})))

(defmulti sanity-check-run
  "Sanity check the given run step"
  (fn [m] (:run m)))
(defmethod sanity-check-run :default sanity-check-run-default
  [m]
  (when (nil? (get-method run (:run m)))
    (throw (ex-info "Cannot dispatch run map" {:m m}))))

(declare dispatch)

(defmethod run :sleep/sleep run-sleep-sleep
  [sys {:sleep/keys [milliseconds]}]
  (info (format "waiting for %d milliseconds" milliseconds))
  (Thread/sleep milliseconds)
  sys)

(defmethod run :process/run run-process-run
  [sys {:process/keys [id cmd port port-timeout file file-timeout wait opts] :or {port-timeout 5000}}]
  (start-proc! sys id cmd opts)
  (when port
    (wait-port sys port {:timeout-ms port-timeout}))
  (when file
    (wait-file file file-timeout))
  (when wait
    (wait-proc! sys id))
  sys)

(defmethod sanity-check-run :process/run sanity-check-process-run
  [{:process/keys [id cmd port port-timeout file file-timeout wait opts] :or {port-timeout 5000} :as m}]
  (when (every? nil? [port file wait])
    (throw (ex-info "must set :process/port, :process/file or :process/wait" {:m m}))))

(defmulti check (fn [sys m] (:check m)))

(defmulti sanity-check-check (fn [m] (:check m)))
(defmethod sanity-check-check :default sanity-check-check-default
  [m]
  (when (nil? (get-method check (:check m)))
    (throw (ex-info "Cannot dispatch check map" {:m m}))))

(defn- report-check
  [{:chk/keys [ok? description] :as chk}]
  (if (:chk/ok? chk)
    (info (format "check succeeded: %s" description) chk)
    (warn (format "check failed: %s" description) chk)))

(defn- loop-single-check
  [sys loop-check set-fail]
  (loop []
    (let [res (check sys loop-check)]
      (set-fail res)
      (if (:chk/ok? res)
        (do
          (report-check res)
          res)
        (do
          (Thread/sleep 1000)
          (recur))))))

(defn- wait-futures
  [end-time futures]
  (mapv (fn [f]
          (let [time-left (- end-time (System/currentTimeMillis))
                res (deref f time-left ::timed-out)]
            (when (= res ::timed-out)
              (future-cancel f))
            res))
        futures))

(defmethod check :loop/forever check-loop-forever
  [sys m]
  (info "waiting forever")
  (deref (promise)))

(defmethod check :loop/until check-loop-until
  [sys {:loop/keys [checks timeout-ms description] :or {description "loop/until"} :as m}]
  (let [stime (System/currentTimeMillis)
        end-time (+ stime timeout-ms)
        last-fails (atom {})
        futures (mapv (fn [loop-check]
                        (future (loop-single-check sys
                                                   loop-check
                                                   (fn [res] (swap! last-fails assoc loop-check res)))))
                      checks)
        results (->> futures
                     (wait-futures end-time)
                     (map (fn [loop-check res]
                            (if (identical? res ::timed-out)
                              {:chk/ok? false
                               :chk/description "timed out"
                               :chk/info (get @last-fails loop-check)}
                              res))
                          checks))
        running-time (- (System/currentTimeMillis) stime)]
    {:chk/ok? (every? :chk/ok? results)
     :chk/running-time running-time
     :chk/description description
     :chk/info results}))

(declare sanity-check-step)
(defmethod sanity-check-check :loop/until sanity-check-check-loop-until
  [{:loop/keys [checks timeout-ms] :as m}]
  (when (not (integer? timeout-ms))
    (throw (ex-info ":loop/timeout-ms must be a number" {:m m})))
  (sanity-check-step checks))

(defn dispatch
  [sys d]
  (if (map? d)
    (cond
      (:run d)
      (update (run sys d) :report report-count-action)
      (:check d)
      (let [check-result (check sys d)]
        (report-check check-result)
        (update sys :report report-add-check-result check-result))
      :else
      (throw (ex-info "cannot dispatch" {:d d})))
    (reduce dispatch sys d)))

(defn- cleanup-processes
  [sys]
  (doseq [[proc-id p]  @*process-map*
          :let [^java.lang.Process proc (:proc p)]]
    (when (.isAlive proc)
      (info "killing process" proc-id)
      (p/destroy-tree p))
    (swap! *process-map* assoc proc-id (deref p))))

(def default-conf {:num-keypers 3, :threshold 2})

(defn- sanity-check-step
  [d]
  (if (map? d)
    (cond
      (:run d)
      (sanity-check-run d)
      (:check d)
      (sanity-check-check d)
      :else
      (throw (ex-info "cannot dispatch" {:d d})))
    (doseq [x d]
      (sanity-check-step x))))

(defn sanity-check-test
  [{:test/keys [id description steps conf]}]
  (sanity-check-step steps))

(defn- sys-run-test
  [{:keys [description cwd] :as sys} {:test/keys [steps] :as tc}]
  (binding [*process-map* (atom {})]
    (try
      (sanity-check-test tc)
      (spit (-> cwd (fs/path "test.edn") str) (puget/pprint-str tc))
      (dispatch sys steps)
      (catch Exception err
        (error err)
        (assoc sys :exception err))
      (finally
        (cleanup-processes sys)))))

(defn- clean-workdir
  "removes all files from the working directory, but keep the directory structure.
  This is useful for those who tail the logfiles and cd to the working directory."
  [root]
  (when (fs/exists? root)
    (fs/walk-file-tree root
                       {:visit-file (fn [path _]
                                      (fs/delete path)
                                      :continue)
                        :post-visit-dir (fn [path _]
                                          :continue)})))

(def ^:dynamic *current-test-id* nil)
(defn run-test
  [{:test/keys [id description conf] :as tc}]
  (binding [*current-test-id* id]
    (let [cwd (-> "work" (fs/path (name id)) str)
          log-dir (-> cwd (fs/path "logs") str)
          log-file (-> cwd (fs/path "test-log.txt") str)
          bb-edn (if (fs/exists? "ci-bb.edn")
                   "ci-bb.edn"
                   "bb.edn")
          bb-edn (-> bb-edn fs/absolutize str)
          sys {:conf (merge default-conf conf)
               :bb-edn bb-edn
               :exception nil
               :id id
               :description description
               :cwd cwd
               :log-dir log-dir
               :report empty-report}]
      (clean-workdir cwd)
      (fs/create-dirs log-dir)
      (binding [play/*cwd* cwd]
        (timbre/with-merged-config
          {:appenders {:test-logs (taoensso.timbre.appenders.core/spit-appender {:fname log-file})}}
          (info (format "Start running test: %s\nYou can start this yourself by running `clojure -M:test %s` inside the play directory."
                        description
                        (name (:test/id tc)))
                tc)
          (let [sys (sys-run-test sys tc)]
            (info (format "Finished test: %s" description) (-> sys :report (dissoc :report/checks)))
            sys))))))

;;; --- Configure logging using puget to pretty print data structures
(defn- indent
  [prefix s]
  (let [lines (str/split s #"\n")]
    (if (= 1 (count lines))
      (str " " s)
      (str/join (interleave (repeat "\n")
                            (repeat prefix)
                            (str/split s #"\n"))))))

(defn- timbre-output-fn
  ([data] (timbre-output-fn nil data))
  ([opts data] ; For partials
   (let [{:keys [no-stacktrace? stacktrace-fonts]} opts
         {:keys [level ?err vargs msg_ ?ns-str ?file hostname_
                 timestamp_ ?line]} data
         [v0 pprint-vargs] (if (string? (first vargs))
                             [(first vargs) (rest vargs)]
                             [nil vargs])]
     (str
      (when-let [ts (force timestamp_)] (str ts " "))
      ;; (force hostname_) " "
      (str/upper-case (name level))  *current-test-id* " "
      "[" (or ?ns-str ?file "?") ":" (or ?line "?") "] - "
      v0
      (when (seq pprint-vargs)
        (indent "    " (puget/cprint-str (if (next pprint-vargs)
                                           pprint-vargs
                                           (first pprint-vargs)))))
      #_(force msg_)
      (when-not no-stacktrace?
        (when-let [err ?err]
          (str enc/system-newline (timbre/stacktrace err opts))))))))

(timbre/merge-config! {:output-fn timbre-output-fn})
