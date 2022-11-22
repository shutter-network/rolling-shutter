(ns sht.collator-test
  (:require [next.jdbc :as jdbc]
            [clojure.string :as str]
            [toml.core :as toml]
            [sht.toml-writer :as toml-writer]
            [taoensso.timbre
             :refer [log  trace  debug  info  warn  error  fatal  report
                     logf tracef debugf infof warnf errorf fatalf reportf
                     spy get-env]]
            [cheshire.core :as json]
            [sht.base64 :as base64]
            [sht.runner :as runner]
            [sht.build :as build]
            [babashka.fs :as fs]
            [sht.play :as play]))

(set! *warn-on-reflection* true)

(defonce play-db-password (or (System/getenv "PLAY_DB_PASSWORD") ""))
(defn- check-query
  [description sys opts query ok-rows?]
  (let [db {:dbtype "postgresql"
            :dbname (play/collator-db)
            :password play-db-password}
        ds (jdbc/get-datasource db)
        query (if (vector? query)
                query
                (query sys opts))
        description (if (string? description)
                      description
                      (description sys opts))
        rows (jdbc/execute! ds query)]
    {:chk/ok? (boolean (ok-rows? sys opts rows))
     :chk/description description
     :chk/info rows}))

(defmacro def-check-query
  [id query ok-rows? description]
  `(defmethod runner/check ~id
     [sys# opts#]
     (check-query ~description sys# opts# ~query ~ok-rows?)))

(defn rows-not-empty?
  [sys opts rows]
  (seq rows))

(defn rows-empty?
  [sys opts rows]
  (empty? rows))

(def-check-query :collator/no-decryption-trigger
  ["select * from decryption_trigger"]
  rows-empty?
  "no decryption trigger should have been generated")

(def-check-query :collator/decryption-trigger
  ["select * from decryption_trigger"]
  rows-not-empty?
  "decryption trigger should have been generated")

(def-check-query :collator/have-batch-tx
  ["select * from batchtx"]
  (fn [sys {:collator/keys [num-batchtxs]} rows]
    (>= (count rows) num-batchtxs))
  (fn [sys {:collator/keys [num-batchtxs]}]
    (format "at least %d batch txs should have been generated" num-batchtxs)))

(defmethod runner/run ::add-collator
  [sys m]
  (let [addr (get-in sys [:sys/collator :subcommand/cfg :eth-address])
        mock-port (get-in sys [:sys/mocksequencer :subcommand/listening-port])
        mock-url (format "http://localhost:%d" mock-port)]
    (play/add-collator mock-url addr 0)
    sys))

(defn test-collator-basic
  [{:keys [num-keypers] :as conf}]
  {:test/id :collator-basic-works
   :test/conf conf
   :test/description "collator basic functionality should work"
   :test/steps [{:run :init/init
                 :init/conf conf}
                (for [keyper (range num-keypers)]
                  {:check :keyper/meta-inf
                   :keyper-num keyper})

                (build/run-chain)
                (build/run-node conf)
                (build/run-keypers conf)
                (build/run-mocksequencer)
                (build/run-collator)
                {:run ::add-collator}

                ;; the keypers are already running, but they won't have a key generated at this
                ;; point. Hence the collator should not generate a decryption trigger.
                {:run :sleep/sleep :sleep/milliseconds 6000}
                {:check :collator/no-decryption-trigger}

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
                                 :keyper/eon 1
                                 :keyper/num keyper})}

                (for [keyper (range num-keypers)]
                  {:check :keyper/non-zero-activation-block
                   :keyper/num keyper})
                {:check :loop/until
                 :loop/description "decryption trigger should be generated"
                 :loop/timeout-ms (* 6 1000)
                 :loop/checks [{:check :collator/decryption-trigger}]}
                {:check :loop/until
                 :loop/description "batchtx generation should work"
                 :loop/timeout-ms (* 20 1000)
                 :loop/checks [{:check :collator/have-batch-tx
                                :collator/num-batchtxs 5}]}
                ;; {:check :loop/forever}
                ]})



(def tests (delay [(test-collator-basic {:num-keypers 3, :threshold 2})]))
