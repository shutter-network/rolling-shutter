(ns sht.core
  (:require [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [sht.runner :as runner]
            [sht.collator-test :as collator-test]
            [sht.dkg-test :as dkg-test])
  (:gen-class))

(defn sanity-check-cases
  [test-cases]
  (doseq [tc test-cases]
    (try
      (runner/sanity-check-test tc)
      (catch Exception err
        (error err "sanity check failed" tc)
        (System/exit 1))))
  test-cases)

(defn report-single-result
  [sys]
  (let [{:report/keys [num-checks-failed num-checks-succeeded checks]} (:report sys)
        failed-checks (remove :chk/ok? checks)
        fail-count-by-description (reduce (fn [m d]
                                            (update m d (fnil inc 0)))
                                          {}
                                          (map :chk/description failed-checks))]
    (println (str (if (zero? num-checks-failed) "  OK" "FAIL")
                  "  "
                  (name (:id sys))
                  ": "
                  (:description sys)))
    (println "     " num-checks-failed "failed," num-checks-succeeded "succeeded")
    (doseq [c (->> failed-checks (map :chk/description) distinct)]
      (printf "      - %s [%dx]\n" c (fail-count-by-description c)))))

(defn sys-succeeded?
  [sys]
  (and (nil? (:exception sys))
       (zero? (-> sys :report :report/num-checks-failed))))

(defn report-result
  [sysv]
  (doseq [sys sysv]
    (report-single-result sys)))

(defn- exit
  [code msg]
  (println msg)
  (System/exit code))

(defn run-test-cases
  [test-cases]
  (let [sysv (mapv runner/run-test test-cases)]
    (println "\n\n=============================================================================\n")
    (report-result sysv)
    (if (every? sys-succeeded? sysv)
      (exit 0 "OK")
      (exit 1 "FAIL"))))

(defn ^:private all-test-cases
  []
  (sanity-check-cases (concat @dkg-test/tests
                              @collator-test/tests)))

(defn run-tests
  [{:keys [nr] :as opts}]
  (let [test-cases (all-test-cases)
        test-cases (if nr [(nth test-cases nr)] test-cases)]
    (run-test-cases test-cases)))

(defn -main
  [& args]
  (let [selected (set args)
        test-cases (all-test-cases)
        test-cases (if (empty? selected)
                     test-cases
                     (filter (comp selected name :test/id) test-cases))]
    (run-test-cases test-cases)))

(comment
  (def sys (runner/run-test (dkg-test/test-keypers-dkg-generation {:num-keypers 3 :num-bootstrappers 2}))))
