const shutter = require("../dist/shutter-crypto");
const ethers = require("ethers");

var encrypted;
describe("Test shutter crypto", () => {
  test("init the wasm library", async () => {
    try {
      await shutter.init();
    } catch (error) {
      expect(error).toBeNone();
    }
  });
  test("encrypt a message", async () => {
    const msg = ethers.getBytes(Buffer.from("a message"));
    const hexString =
      "0b4e86e0ed51ef774210d1c0fe0be6f1b4f0695d5d396b3b003547f752ac82e316375aa37b1739c9c8472b1b5ae09477565bf9d2c0d7db0c39576f4615d32703262d5854bfbac8a60eb6d227f397289e6e51f979b56476b7f7f32a45ede7a61f21d893a54ab6e65b283342adc41d53df5432569c6a8c2304921bce3ea148efb4";
    const eonkey = Uint8Array.from(Buffer.from(hexString, "hex"));
    const epochId = ethers.getBytes(
      ethers.zeroPadValue(Buffer.from("46", "hex"), 32)
    );

    const sigma = new Uint8Array(32);
    crypto.getRandomValues(sigma);
    try {
      encrypted = await shutter.encrypt(msg, eonkey, epochId, sigma);
    } catch (error) {
      expect(error).toBeNull();
    }
    expect(encrypted).toBeDefined();
  });
  test("decrypt a message", async () => {
    const decryptionKey = Uint8Array.from(
      Buffer.from(
        "009bb51574d6a6790faa4724dfad416ca059a286ccfee20be732cac9a81e05dc2f47905cbaa0fb043ff849b0c41e99208d98d27cba3fffb43d63ba50c35259d3",
        "hex"
      )
    );
    var decrypted;
    try {
      decrypted = await shutter.decrypt(encrypted, decryptionKey);
    } catch (error) {
      expect(error).toBeNull();
    }
  });
});

describe("Test known values (values obtained from 'rolling-shutter crypto encrypt/decrypt')", () => {
  test("encrypt a message with zero sigma", async () => {
    const epochId = ethers.getBytes(
      ethers.zeroPadValue(Buffer.from("46", "hex"), 32)
    );
    const eonKey = ethers.getBytes(
      Buffer.from(
        "0b4e86e0ed51ef774210d1c0fe0be6f1b4f0695d5d396b3b003547f752ac82e316375aa37b1739c9c8472b1b5ae09477565bf9d2c0d7db0c39576f4615d32703262d5854bfbac8a60eb6d227f397289e6e51f979b56476b7f7f32a45ede7a61f21d893a54ab6e65b283342adc41d53df5432569c6a8c2304921bce3ea148efb4",
        "hex"
      )
    );
    const sigma = ethers.getBytes(
      Buffer.from(
        "0000000000000000000000000000000000000000000000000000000000000000",
        "hex"
      )
    );
    const message = Uint8Array.from(
      "a message".split("").map((c) => c.charCodeAt())
    );
    const expected_encrypted =
      "0x01f2490511e502db0ea9940cfc16ae5ee4435e70b8f9a5567c6230ee41026d5710cdeeaef09dbed3d461592995150eaafba0ed3eeef5be914172677fa9a095261cc049a326159a7ad35b80ba3c08296bef0c19fc6e605e0c21d542a46e83eb611ee04582022da7715b417c91502f07e26eb642f20918830a08f584afb04cb3b0b88d48b4aa8d6e5cf3e6066f4b765774d0ff84d046ef6c92f6c3ad1a2711aecf92baebfa11943d9295a886519fb1a59edde5138322cff19b4b497a12f35e04e4";
    const result = await shutter.encrypt(message, eonKey, epochId, sigma);
    const result_hex = ethers.hexlify(result);
    expect(result_hex).toEqual(expected_encrypted);
  });
  test("decrypt a known encrypted message", async () => {
    const known_encrypted = ethers.getBytes(
      Buffer.from(
        "01f2490511e502db0ea9940cfc16ae5ee4435e70b8f9a5567c6230ee41026d5710cdeeaef09dbed3d461592995150eaafba0ed3eeef5be914172677fa9a095261cc049a326159a7ad35b80ba3c08296bef0c19fc6e605e0c21d542a46e83eb611ee04582022da7715b417c91502f07e26eb642f20918830a08f584afb04cb3b0b88d48b4aa8d6e5cf3e6066f4b765774d0ff84d046ef6c92f6c3ad1a2711aecf92baebfa11943d9295a886519fb1a59edde5138322cff19b4b497a12f35e04e4",
        "hex"
      )
    );
    const decryptionKey = Uint8Array.from(
      Buffer.from(
        "009bb51574d6a6790faa4724dfad416ca059a286ccfee20be732cac9a81e05dc2f47905cbaa0fb043ff849b0c41e99208d98d27cba3fffb43d63ba50c35259d3",
        "hex"
      )
    );
    var decrypted;
    try {
      decrypted = await shutter.decrypt(known_encrypted, decryptionKey);
    } catch (error) {
      expect(error).toBeNull();
    }
    const expected = Uint8Array.from(
      "a message".split("").map((c) => c.charCodeAt())
    );
    expect(decrypted).toEqual(expected);
  });
});
