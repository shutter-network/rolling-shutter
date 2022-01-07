(ns sht.core
  (:require [next.jdbc :as jdbc]
            [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [sht.runner :as runner]
            [sht.build :as build]
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
   :test/steps [(build/init conf)
                (for [keyper (range num-keypers)]
                  {:check :keyper/meta-inf
                   :keyper-num keyper})

                {:run :process/run
                 :process/id :chain
                 :process/cmd '[bb chain]
                 :process/port 26657}

                {:run :process/run
                 :process/id :node
                 :process/cmd '[bb node]
                 :process/port 8545
                 :process/port-timeout (+ 5000 (* num-keypers 2000))}

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

(defn test-dkg-keypers-join-late
  [{:keys [num-keypers threshold] :as conf}]
  {:test/id :keyper-dkg-works-keyper-joins-late
   :test/conf conf
   :test/description "distributed key generation should work when a keyper joins late"
   :test/steps [(build/init conf)
                (for [keyper (range num-keypers)]
                  {:check :keyper/meta-inf
                   :keyper-num keyper})

                {:run :process/run
                 :process/id :chain
                 :process/cmd '[bb chain]
                 :process/port 26657}

                {:run :process/run
                 :process/id :node
                 :process/cmd '[bb node]
                 :process/port 8545
                 :process/port-timeout (+ 5000 (* num-keypers 2000))}

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
  (let [{:report/keys [num-actions num-checks-failed num-checks-succeeded checks]} (:report sys)
        failed-checks (remove :chk/ok? checks)]
    (println (str (if (zero? num-checks-failed) "  OK" "FAIL")
                  "  "
                  (name (:id sys))
                  ": "
                  (:description sys)))
    (println "     " num-checks-failed "failed," num-checks-succeeded "succeeded")
    (doseq [c failed-checks]
      (println "     " c))))

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

(defn run-tests
  [opts]
  (let [test-cases (sanity-check-cases (generate-tests))
        sysv (mapv runner/run-test test-cases)]
    (report-result sysv)
    (if (every? sys-succeeded? sysv)
      (exit 0 "OK")
      (exit 1 "FAIL"))))

(comment
  (def sys (runner/run-test (test-keypers-dkg-generation {:num-keypers 3, :num-decryptors 2})))
  )
