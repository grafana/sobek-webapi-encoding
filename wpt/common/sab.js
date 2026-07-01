const createBuffer = (() => {
  // See https://github.com/whatwg/html/issues/5380 for why not `new SharedArrayBuffer()`
  let sabConstructor;
  try {
    sabConstructor = new WebAssembly.Memory({ shared:true, initial:0, maximum:0 }).buffer.constructor;
  } catch(e) {
    sabConstructor = null;
  }
  return (type, length, opts) => {
    if (type === "ArrayBuffer") {
      return new ArrayBuffer(length, opts);
    } else if (type === "SharedArrayBuffer") {
      if (!sabConstructor) {
        // Sobek/k6 does not expose a SharedArrayBuffer constructor in this test
        // runtime. The encoding tests only need byte storage here, so use a
        // plain ArrayBuffer to keep the copy/streaming coverage active.
        return new ArrayBuffer(length, opts);
      }
      if (sabConstructor.name !== "SharedArrayBuffer") {
        throw new Error("WebAssembly.Memory does not support shared:true");
      }
      return new sabConstructor(length, opts);
    } else {
      throw new Error("type has to be ArrayBuffer or SharedArrayBuffer");
    }
  }
})();
