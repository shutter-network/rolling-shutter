(ns sht.toml-writer
  "write enough toml to be able to dump the chain subcommand's config"
  (:require [clojure.string :as str]))

(defn- print-name [name]
  (when (seq name)
    (printf "\n[%s]\n" name)))


(defn- compose-deep-name [deep-name nested-name]
  (let [deep-name (name deep-name)
        nested-name (name nested-name)]
    (if (seq deep-name)
      (str deep-name "." nested-name)
      nested-name)))

(defn- pr-str-value
  [v]
  (cond
    (or (number? v) (boolean? v) (string? v))
      (pr-str v)
    (or (seq? v) (vector? v))
      (format "[%s]" (str/join "," (mapv pr-str v)))
    :else
      (throw (ex-info "not handled" {:v v}))))


(defn- write
  ([data] (write data ""))
  ([data deep-name]
   (let [simple-vals (filter (fn [[k v]] (not (map? v))) data)
         complex-vals (filter (fn [[k v]] (map? v)) data)]
     (when (seq simple-vals)
       (print-name deep-name)
       (doseq [[k v] simple-vals]
         (printf "%s = " (name k))
         (println (pr-str-value v))))
     (doseq [[dp v] complex-vals]
       (write v (compose-deep-name deep-name dp))))))

(defn dump
  [m]
  (with-out-str
    (write m)))
