const shutter = require("../dist/shutter-crypto");

describe("Sum numbers", () => {
  test("it should sum two numbers correctly", () => {
    const sum = 1 + 2;
    const expectedResult = 3;
    expect(sum).toEqual(expectedResult);
  });
});

describe("Test shutter crypto", () => {
  test("init the wasm library", async () => {
    try {
      await shutter.init();
    } catch (error) {
      expect(error).toBeNone();
    }
    const msg = new Uint8Array(32);
    const hexString =
      "0x0b4e86e0ed51ef774210d1c0fe0be6f1b4f0695d5d396b3b003547f752ac82e316375aa37b1739c9c8472b1b5ae09477565bf9d2c0d7db0c39576f4615d32703262d5854bfbac8a60eb6d227f397289e6e51f979b56476b7f7f32a45ede7a61f21d893a54ab6e65b283342adc41d53df5432569c6a8c2304921bce3ea148efb4";
    const eonkey = Uint8Array.from(Buffer.from(hexString, "hex"));
    const epochId = 0x46;
    const sigma = new Uint8Array(32);

    const encrypted = shutter.encrypt(msg, eonkey, sigma);
    // expect(encrypted).toBeDefined();
    const decryptionKey =
      "0x009bb51574d6a6790faa4724dfad416ca059a286ccfee20be732cac9a81e05dc2f47905cbaa0fb043ff849b0c41e99208d98d27cba3fffb43d63ba50c35259d3";
  });
});
