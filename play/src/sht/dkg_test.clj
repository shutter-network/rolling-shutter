(ns sht.dkg-test
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
