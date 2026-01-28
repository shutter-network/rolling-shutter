#!/usr/bin/env bb

(require '[babashka.classpath :as cp]
         '[babashka.process :as p]
         '[babashka.cli :as cli]
         '[babashka.fs :as fs])

(let [src-dir (str (fs/normalize (fs/absolutize (fs/path *file* "../../play/src"))))]
  (cp/add-classpath src-dir))

(require '[sht.ninjagen :as n])

(def spec
  {:gen
   {:desc "Whether to run code generators"
    :default true}

   :install-gen
   {:desc "Whether to install code generators"
    :default true}

   :run-ninja
   {:desc "Run ninja"
    :alias :x}

   :help
   {:desc "Show help message"
    :alias :h}})

(defn print-help
  []
  (println "Usage: build.clj [options]")
  (println (cli/format-opts {:spec spec :order [:gen :install-gen :run-ninja :help]}))
  (System/exit 0))

(defn -main [& args]
  (let [{:keys [help gen install-gen run-ninja]} (cli/parse-opts args
                                                                 {:spec spec
                                                                  :restrict (keys spec)})
        shutter-root (n/absnormpath *file* "..")
        build-ninja (n/absnormpath *file* ".." "build.ninja")]
    (when help
      (print-help))
    ;; (println "Generating"  build-ninja)
    (-> {:shutter-root shutter-root
         :install-code-generators? install-gen
         :run-code-generators? gen
         :dst-dir (n/absnormpath *file* "../bin")}
        n/configure
        (n/gen-file build-ninja))
    (when run-ninja
      (p/check (p/process {:out :inherit :err :inherit :in :inherit}
                          "ninja" "-C" shutter-root)))))

(when (= *file* (System/getProperty "babashka.file"))
  (apply -main *command-line-args*))
