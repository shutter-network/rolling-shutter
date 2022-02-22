(ns sht.base64)

(defn encode
  "Encode an array of bytes into a base64 encoded string."
  [^bytes unencoded]
  (String. (.encode (java.util.Base64/getEncoder) unencoded)))

(defn decode
  "Decode a base64 encoded string into an array of bytes."
  [^String encoded]
  (.decode (java.util.Base64/getDecoder) encoded))
