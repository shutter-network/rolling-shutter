;; this namespace is read by babashka's via our bb.edn file. Make sure it stays compatible with
;; babashka
(ns sht.play
  (:require [clojure.edn :as edn]
            [clojure.java.io :as io]
            [clojure.pprint :as pprint]))

(defn keyper-db
  [n]
  (format "keyper-db-%d" n))

(defn decryptor-db
  [n]
  (format "decryptor-db-%d" n))

(defn ci-gen
  "Rewrite bb.edn with a simplified build for use on CI systems"
  []
  (let [src "bb.edn"
        dst "ci-bb.edn"
        bb (edn/read-string (slurp src))
        bb (assoc-in bb [:tasks 'build :depends] ['-go-files])]
    (with-open [w (io/writer dst)]
      (pprint/pprint bb w))
    (println "Created simpified config in" dst)))
