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
                ;; {:check :loop/forever}
                ]})



(def tests (delay [(test-collator-basic {:num-keypers 3, :threshold 2})]))
