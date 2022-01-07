(ns sht.build
  "sht.build contains some functions that build steps for use in tests")

(defn init
  [{:keys [num-keypers num-decryptors]}]
  [{:run :process/run
    :process/id :build
    :process/cmd '[bb build]
    :process/wait true}
   {:run :process/run
    :process/id :init
    :process/cmd '[bb init]
    :process/wait true}])

(defn run-keyper
  [n]
  {:run :process/run
   :process/id (keyword (format "keyper-%d" n))
   :process/cmd ['bb "k" (str n)]})

(defn run-keypers
  [{:keys [num-keypers]}]
  (mapv run-keyper (range num-keypers)))
