/**
 * Jest Test Setup
 * Configures global testing environment for headless testing with real WASM
 */

import * as fs from 'fs';
import * as path from 'path';
import { TextEncoder, TextDecoder } from 'util';

// Polyfill TextEncoder/TextDecoder for Node.js environment
(global as any).TextEncoder = TextEncoder;
(global as any).TextDecoder = TextDecoder;

// Set up real WASM environment for Node.js
const wasmPath = path.join(__dirname, '../static/wasm/lilbattle-cli.wasm');
const wasmExecPath = path.join(__dirname, '../static/wasm/wasm_exec.js');

// Load Go's WASM exec helper
if (fs.existsSync(wasmExecPath)) {
  const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
  eval(wasmExecCode);
  
  // Make sure Go is available on window for GameState
  (window as any).Go = (global as any).Go;
}

// Implement fetch for WASM loading in Node.js environment
global.fetch = jest.fn().mockImplementation((url: string) => {
  console.log('Fetch called with URL:', url);
  
  // Handle WASM file requests
  if (url.includes('lilbattle') && url.includes('.wasm')) {
    console.log('WASM fetch detected, checking path:', wasmPath);
    if (fs.existsSync(wasmPath)) {
      const wasmBuffer = fs.readFileSync(wasmPath);
      // Create a proper Response-like object that matches Web API
      const response = {
        arrayBuffer: () => Promise.resolve(wasmBuffer.buffer.slice(
          wasmBuffer.byteOffset, 
          wasmBuffer.byteOffset + wasmBuffer.byteLength
        )),
        ok: true,
        status: 200,
        statusText: 'OK',
        headers: new Map(),
        body: null,
        bodyUsed: false,
        clone: function() { return this; },
        json: () => Promise.reject(new Error('Not JSON')),
        text: () => Promise.reject(new Error('Not text')),
        blob: () => Promise.reject(new Error('Not blob')),
        formData: () => Promise.reject(new Error('Not form data'))
      };
      console.log('Returning WASM response');
      return Promise.resolve(response);
    } else {
      console.error(`WASM file not found at ${wasmPath}`);
      return Promise.reject(new Error(`WASM file not found at ${wasmPath}`));
    }
  }
  
  // Handle other requests - return empty successful response
  console.log('Non-WASM fetch, returning empty response');
  const emptyResponse = {
    ok: true,
    status: 200,
    statusText: 'OK',
    headers: new Map(),
    body: null,
    bodyUsed: false,
    clone: function() { return this; },
    arrayBuffer: () => Promise.resolve(new ArrayBuffer(0)),
    json: () => Promise.resolve({}),
    text: () => Promise.resolve(''),
    blob: () => Promise.resolve(new Blob()),
    formData: () => Promise.resolve(new FormData())
  };
  return Promise.resolve(emptyResponse);
});

// Mock WebAssembly.instantiateStreaming for Node.js
if (!global.WebAssembly.instantiateStreaming) {
  global.WebAssembly.instantiateStreaming = async (response: any, importObject: any) => {
    const arrayBuffer = await response.arrayBuffer();
    console.log("Here?????")
    const out = WebAssembly.instantiate(arrayBuffer, importObject);
    console.log("2. Here?????")
    return out
  };
}

// Mock DOM APIs that Phaser might use
Object.defineProperty(window, 'requestAnimationFrame', {
  value: jest.fn((cb) => setTimeout(cb, 16))
});

Object.defineProperty(window, 'cancelAnimationFrame', {
  value: jest.fn()
});

// Mock Phaser for headless testing - prevent initialization issues
jest.mock('phaser', () => ({
  Scene: class MockScene {
    constructor() {}
    create() {}
    preload() {}
    update() {}
  },
  Game: class MockGame {
    constructor() {}
    destroy() {}
  },
  GameObjects: {
    Graphics: class MockGraphics {},
    Text: class MockText {},
    Sprite: class MockSprite {},
  },
  Types: {
    Input: {
      Keyboard: {}
    }
  }
}));

// Mock Canvas and WebGL context for headless testing  
HTMLCanvasElement.prototype.getContext = jest.fn((contextId: string) => {
  if (contextId === '2d') {
    return {
      fillStyle: '',
      strokeStyle: '',
      lineWidth: 1,
      fillRect: jest.fn(),
      strokeRect: jest.fn(),
      clearRect: jest.fn(),
      beginPath: jest.fn(),
      moveTo: jest.fn(),
      lineTo: jest.fn(),
      closePath: jest.fn(),
      stroke: jest.fn(),
      fill: jest.fn(),
      save: jest.fn(),
      restore: jest.fn(),
      scale: jest.fn(),
      translate: jest.fn(),
      rotate: jest.fn(),
      drawImage: jest.fn(),
      createImageData: jest.fn(),
      getImageData: jest.fn(),
      putImageData: jest.fn(),
    };
  }
  if (contextId === 'webgl' || contextId === 'webgl2') {
    return {
      clearColor: jest.fn(),
      clear: jest.fn(),
      clearDepth: jest.fn(),
      enable: jest.fn(),
      disable: jest.fn(),
      getParameter: jest.fn(),
      getExtension: jest.fn(),
      createShader: jest.fn(),
      shaderSource: jest.fn(),
      compileShader: jest.fn(),
      createProgram: jest.fn(),
      attachShader: jest.fn(),
      linkProgram: jest.fn(),
      useProgram: jest.fn(),
      getAttribLocation: jest.fn(),
      getUniformLocation: jest.fn(),
      createBuffer: jest.fn(),
      bindBuffer: jest.fn(),
      bufferData: jest.fn(),
      enableVertexAttribArray: jest.fn(),
      vertexAttribPointer: jest.fn(),
      drawArrays: jest.fn(),
      drawElements: jest.fn(),
      createTexture: jest.fn(),
      bindTexture: jest.fn(),
      texImage2D: jest.fn(),
      texParameteri: jest.fn(),
      generateMipmap: jest.fn(),
    };
  }
  return null;
}) as any;

// Console setup for better test output
const originalError = console.error;
beforeAll(() => {
  // Suppress expected error messages during testing
  console.error = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      (args[0].includes('WebGL') || 
       args[0].includes('WASM') ||
       args[0].includes('Failed to load'))
    ) {
      return;
    }
    originalError.call(console, ...args);
  };
});

afterAll(() => {
  console.error = originalError;
});
