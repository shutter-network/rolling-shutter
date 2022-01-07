;; this namespace is read by babashka's via our bb.edn file. Make sure it stays compatible with
;; babashka
(ns sht.play)

(defn keyper-db
  [n]
  (format "keyper-db-%d" n))

(defn decryptor-db
  [n]
  (format "decryptor-db-%d" n))
