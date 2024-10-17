// jest hackery: https://github.com/mswjs/msw/discussions/1934
const { ReadableStream } = require('node:stream/web');
Object.defineProperties(globalThis, {
  ReadableStream: { value: ReadableStream },
});

export {};
