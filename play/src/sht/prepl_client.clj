(ns sht.prepl-client
  (:import (java.net Socket)
           (java.io DataInputStream DataOutputStream) ))

(def IPaddress "127.0.0.1")
(def port 1666)
(def command "(+ 1 1)\n")
(def socket (Socket. IPaddress port))

(def in (DataInputStream. (.getInputStream socket)))
(def out (DataOutputStream. (.getOutputStream socket)))

(println "Input:" command)
(.writeUTF out command)

(def response (.readLine in))
(println "Output: " response)
