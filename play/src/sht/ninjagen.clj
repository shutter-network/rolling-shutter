(ns sht.ninjagen
  (:require [clojure.string :as str]
            [babashka.fs :as fs]
            [babashka.process :as p]))

(defn- indent
  [lines]
  (map (partial str "  ") lines))

(defn- gen-vars
  [m]
  (->> m
       (map (fn [[k v]]
              [(name k) (str v)]))
       sort
       (mapv (fn [[k v]]
               (format "%s = %s" k v)))))

(defn gen
  [ds]
  (cond
    (map? ds)
      (gen-vars ds)
    (string? ds)
      [ds]
    (nil? ds)
      []
    :else
      (mapcat gen ds)))

(defn- format-path
  [p]
  (str/replace (str p) " " "$ "))

(defn- format-paths
  [ps]
  (if (string? ps)
    (format-path ps)
    (str/join " " (map format-path ps))))

(defn rule
  [rule-name & {:as vars}]
  [(str "rule " rule-name)
   (indent (gen-vars vars))
   ""])

(defn build
  [rule & {:keys [outputs inputs implicit-deps vars]}]
  [(str "build " (format-paths outputs) ": "
        (name rule) " " (format-paths inputs)
        (when (seq implicit-deps) (str " | " (format-paths implicit-deps))))
   (indent (gen vars))
   ""])

(defn gen-file
  [ds path]
  (let [lines (remove nil? (gen ds))
        ninja-str (str/join (interleave lines (repeat \newline)))]
    (spit path ninja-str)))

(def absnormpath (comp str fs/absolutize fs/normalize fs/path))

(def rule-npm-install
  (rule "npm-install" {:command "cd $src-dir; npm install; touch $mark-path"
                       :description "Running npm install in $src-dir"}))

(defn build-npm-project
  [src-dir & {:keys [alias]}]
  (let [implicit-deps [(absnormpath src-dir "package.json")
                       (absnormpath src-dir "package-lock.json")]
        mark-path (absnormpath src-dir "node_modules/.mark-npm-install")]
    [(build "npm-install"
            :outputs [mark-path]
            :implicit-deps implicit-deps
            :vars {:src-dir src-dir
                   :mark-path mark-path})
     (when-not (empty? alias)
       (build "phony" :outputs [alias] :inputs [mark-path]))]))


(defn go-files
  [src-dir]
  (->> (p/process ["go" "list" "-m" "-f={{.Dir}}"]
                  {:out :string
                   :err :string
                   :dir src-dir})
       p/check
       :out
       str/split-lines
       (mapcat (fn [dir]
                 (concat (map str (fs/glob dir "go.{mod,sum}"))
                         (sort (map str (fs/glob dir "**.go"))))))))

(def go-code-generators
  ["github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.1"
   "github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0"
   "google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0"
   "github.com/abice/go-enum@v0.5.6"
   "github.com/ethereum/go-ethereum/cmd/abigen@v1.12.0"])

(def rule-oapi-codegen
  ["# Run oapi-codegen"
   (rule "oapi-codegen"
         :command "$oapi-codegen --generate $generate --package $package -o $out $in")])

(def rule-sqlc
  ["# Run sqlc generate"
   (rule "sqlc"
         :command "$sqlc generate --file $in")])

(def rule-protoc
  ["# Run protoc"
   (rule "protoc"
         :command "$protoc --plugin $protoc-gen-go -I${go-out} $basename --go_out ${go-out}")])

(defn build-oapi-codegen
  [& {:keys [package inputs outputs generate implicit-deps]}]
  (build "oapi-codegen"
         :inputs inputs
         :outputs outputs
         :implicit-deps implicit-deps
         :vars {:package package
                :generate generate}))

(defn- pkg-exe-name
  [pkg]
  (second (re-find #"/([^/@]+)(@[^/]*)?$", pkg)))

(def rule-go-install
  (rule "go-install"
        :command "env GOBIN=${dst-dir} go install $pkg"
        :description "Install $pkg to $dst-dir"))

(defn build-go
  [src-dir & {:keys [dst-dir executables subdirs alias]}]
  (let [executables (mapv (fn [e] (str (fs/path dst-dir e))) executables)]
    [(build "go-build"
            :outputs executables
            :implicit-deps (map str (go-files src-dir))
            :vars {:src-dir src-dir
                   :dst-dir dst-dir
                   :subdirs (str/join " " subdirs)})
     (when-not (empty? alias)
       (build "phony" :outputs [alias] :inputs executables))]))

(def go-rules
  ["# Build go projects"
   (rule "go-build"
         :command "$go build -C $src-dir -o $dst-dir $subdirs")
   "# Install go package"
   (rule "go-install"
         :command "env GOBIN=${dst-dir} go install $pkg"
         :description "Install $pkg to $dst-dir")])

(defn go-install
  [pkg & {:keys [dst-dir]}]
  (build "go-install" :outputs [(str (fs/path dst-dir (pkg-exe-name pkg)))]
         :vars {:pkg pkg
                :dst-dir (absnormpath dst-dir)}))

(defn build-sqlc
  [path & {:keys [implicit-deps]}]
  (let [path (fs/absolutize path)
        ;;yml-data (slurp-yml (str path))
        sql-files (fs/glob (fs/parent path) "**.sql")]
    (build "sqlc"
           :outputs [(fs/normalize (fs/path path "../../models.sqlc.gen.go"))]
           :inputs [(str path)]
           :implicit-deps (concat implicit-deps
                                  (sort (map str sql-files))))))

(defn build-protoc
  [path & {:keys [implicit-deps]}]
  (let [dir (str (fs/parent path))]
    (build "protoc"
           :outputs [(str (fs/strip-ext path) ".pb.go")]
           :inputs (str path)
           :implicit-deps implicit-deps
           :vars {:go-out dir
                  :basename (fs/file-name path)})))

(defn configure-oapi
  [{:keys [shutter-root install-code-generators? run-code-generators? dst-dir]}]
  (let [oapi-codegen (if install-code-generators? (absnormpath dst-dir "oapi-codegen") "oapi-codegen")

        implicit-deps (if install-code-generators? [(absnormpath oapi-codegen)] [])]
    ["#"
     "# --- oapi-codegen"
     "#"
     {:oapi-codegen oapi-codegen}
     ""
     rule-oapi-codegen
     (build-oapi-codegen :package "client"
                         :generate "types,client"
                         :inputs (absnormpath shutter-root "collator/oapi/oapi.yaml")
                         :outputs (absnormpath shutter-root "collator/client/client.gen.go")
                         :implicit-deps implicit-deps)
     (build-oapi-codegen :package "oapi"
                         :generate "types,chi-server,spec"
                         :inputs (absnormpath shutter-root "collator/oapi/oapi.yaml")
                         :outputs (absnormpath shutter-root "collator/oapi/oapi.gen.go")
                         :implicit-deps implicit-deps)
     (build-oapi-codegen :package "kproapi"
                         :generate "types,chi-server,spec"
                         :inputs (absnormpath shutter-root "keyper/kproapi/oapi.yaml")
                         :outputs (absnormpath shutter-root "keyper/kproapi/oapi.gen.go")
                         :implicit-deps implicit-deps)]))

(defn configure-abigen
  [{:keys [shutter-root install-code-generators? dst-dir]}]
  (let [contracts-dir (absnormpath shutter-root "../contracts")
        sol-files (fs/glob contracts-dir "src/**/*.sol")
        abigen (if install-code-generators? (absnormpath dst-dir "abigen") "abigen")
        abigen-js (fs/path contracts-dir "scripts" "abigen.js")
        command (if install-code-generators?
                  (str "env PATH=\"" (absnormpath dst-dir) ":$$PATH\" node " abigen-js)
                  (str "node " abigen-js))
        implicit-deps (if install-code-generators? [(absnormpath abigen)] [])
        implicit-deps (concat implicit-deps
                              ["contracts" abigen-js])]
    ["# --- abigen"
     ""
     {:abigen abigen}
     ""
     (rule "abigen"
           :command command
           :description "Running abigen")
     ""
     (build "abigen"
            :outputs [(fs/path contracts-dir "combined.json")
                      (fs/path shutter-root "contract/binding.abigen.gen.go")]
            :inputs sol-files
            :implicit-deps implicit-deps)

     ]
    ))

(defn configure-sqlc
  [{:keys [shutter-root install-code-generators? dst-dir]}]
  (let [sqlc-yaml-paths (fs/glob shutter-root "**/sqlc.yaml")
        sqlc (if install-code-generators? (absnormpath dst-dir "sqlc") "sqlc")
        implicit-deps (if install-code-generators? [sqlc] [])]
    ["# --- sqlc"
     ""
     {:sqlc sqlc}
     ""
     rule-sqlc
     (for [path sqlc-yaml-paths]
       (build-sqlc path :implicit-deps implicit-deps))]))

(defn configure-protoc
  [{:keys [shutter-root install-code-generators? dst-dir]}]
  (let [proto-paths (fs/glob shutter-root "**/*.proto")
        protoc-gen-go (if install-code-generators? (absnormpath dst-dir "protoc-gen-go") (fs/which "protoc-gen-go"))
        implicit-deps (if install-code-generators? [protoc-gen-go] [])]
    ["# --- protoc"
     ""
     {:protoc "protoc"
      :protoc-gen-go protoc-gen-go}

     rule-protoc
     (for [path proto-paths]
       (build-protoc path :implicit-deps implicit-deps))]))

(defn configure-code-generators
  [{:keys [dst-dir]}]
  (for [pkg go-code-generators]
    (go-install pkg :dst-dir dst-dir)))

(defn configure-rolling-shutter
  [{:keys [shutter-root dst-dir]}]
  ["# --- rolling-shutter"
   (build-go (str shutter-root)
             :dst-dir dst-dir
             :executables ["rolling-shutter" "keygen"]
             :subdirs ["." "./sandbox/keygen"]
             :alias "rolling-shutter")
   ;; (build "phony" :outputs ["rolling-shutter"] :inputs [(absnormpath dst-dir "rolling-shutter")])
   "default rolling-shutter"
   ""])

(defn configure-contracts
  [env]
  (build-npm-project (absnormpath (:shutter-root env) "../contracts") :alias "contracts"))

(defn dump-env
  [env]
  (map (fn [[k v]] (format  "# %28s  %s" k v)) env))

(defn configure
  [{:keys [install-code-generators? run-code-generators? dst-dir] :as env}]
  ["# Please do not edit this file"
   "#"
   (dump-env env)
   ""
   {:go "go"
    :dst-dir dst-dir}
   ""
   go-rules

   (when run-code-generators?
     [(configure-oapi env)
      (configure-sqlc env)
      (configure-protoc env)
      (configure-abigen env)])

   (when install-code-generators?
     (configure-code-generators env))

   (configure-rolling-shutter env)

   rule-npm-install
   (configure-contracts env)

   "# build.ninja ends here"])
