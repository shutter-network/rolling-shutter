(ns gen
  (:require [clj-yaml.core :as yaml]
            [clojure.java.shell :as shell]
            [clojure.string :as str]))

(def default-base "origin/main")

(def rx-table
  "maps regular expressions to keywords, which we use in get-yml-files"
  [[#"contracts/.*" :contracts-changed?]
   [#"\.circleci/contracts.*" :contracts-changed?]

   [#"rolling-shutter/.*" :rolling-shutter-changed?]
   [#"\.circleci/rolling-shutter.*$" :rolling-shutter-changed?]

   [#"\.circleci/basis.yml$" :build-all?]])

(defn extra-keywords
  [{:keys [head base] :as opts}]
  (merge {}
         (when (or (not (empty? (System/getenv "CIRCLE_TAG")))
                   (= "main" (System/getenv "CIRCLE_BRANCH")))
           {:build-all? true})))

(defn get-yml-files
  [{:keys [contracts-changed? rolling-shutter-changed? build-all?] :as kws}]
  (println "Choose yml files based on" kws)
  (remove nil?
          ["basis.yml"
           (when (or contracts-changed? build-all?)
             "contracts.yml")
           (when (or rolling-shutter-changed? build-all?)
             "rolling-shutter.yml")]))

(defn slurp-yml
  "read yaml file from the given path"
  [path]
  (println "Reading" path)
  (yaml/parse-string (slurp path)))

(defn spit-yml
  [m output-path]
  (let [yaml-string (yaml/generate-string m)]
    (with-open [f (clojure.java.io/writer output-path)]
      (spit f yaml-string))))

(defn shell-out
  [& args]
  (let [{:keys [exit out err] :as res} (apply shell/sh args)]
    (if-not (zero? exit)
      (throw (ex-info "running shell command failed" (assoc res :args args))))
    out))

(defn find-changed-files
  [rev1 rev2]
  (str/split-lines (shell-out "git" "diff" "--name-only" rev1 rev2)))

(defn filter-matches
  [coll rx]
  (filter (partial re-matches rx) coll))

(defn match-changed-files
  [changed]
  (reduce (fn [m [rx kw]]
            (if (get m kw)
              m
              (assoc m kw (not (empty? (filter-matches changed rx))))))
          {}
          rx-table))

(defn deep-merge
  [& maps]
  (if (every? map? maps)
    (apply merge-with deep-merge maps)
    (last maps)))

(defn set-head
  [{:keys [head base] :as opts}]
  (let [head (first (remove empty? [head (System/getenv "CIRCLE_SHA1") "HEAD"]))
        head (str/trim (shell-out "git" "rev-parse" head))]
    (assoc opts :head head)))

(defn set-merge-base
  [{:keys [head base] :as opts}]
  (let [base (if (empty? base) default-base base)
        base (str/trim (shell-out "git" "merge-base" base head))]
    (assoc opts :base base)))

(defn build-config
  ([{:keys [base head] :as opts}]
   (let [changed-files (find-changed-files base head)
         kwmatch (match-changed-files changed-files)
         yaml-files (get-yml-files (merge kwmatch (extra-keywords opts)))
         ms (mapv slurp-yml yaml-files)]
     (apply deep-merge ms))))

(def default-options {:validate true
                      :output "continue-generated.yml"})
(defn gen
  "entry point, will be called via clojure -X gen/gen from CircleCI"
  [opts]
  (let [opts (merge default-options opts)
        output (:output opts)]
    (-> opts set-head set-merge-base build-config (spit-yml output))
    (println "Config generated in" output)
    (when (:validate opts)
      (println (shell-out "circleci" "config" "validate" output)))))
