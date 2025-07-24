/**
 * WASM Loading Debug Test
 * Simpler test to understand WASM loading issues
 */

import * as fs from 'fs';
import * as path from 'path';

const STATIC_BASE_PATH = "../static"

describe('WASM Loading Debug', () => {
  
  test('should find WASM files', () => {
    const wasmPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/weewar-cli.wasm');
    const wasmExecPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/wasm_exec.js');
    
    console.log('WASM path:', wasmPath);
    console.log('WASM exec path:', wasmExecPath);
    
    expect(fs.existsSync(wasmPath)).toBe(true);
    expect(fs.existsSync(wasmExecPath)).toBe(true);
    
    const wasmStats = fs.statSync(wasmPath);
    console.log('WASM file size:', wasmStats.size, 'bytes');
  });
  
  test('should load wasm_exec.js successfully', () => {
    const wasmExecPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/wasm_exec.js');
    const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
    
    expect(wasmExecCode).toContain('Go');
    expect(wasmExecCode.length).toBeGreaterThan(1000);
    
    // Try to evaluate the script
    eval(wasmExecCode);
    
    expect(typeof (global as any).Go).toBe('function');
  });
  
  test('should create Go instance', () => {
    const wasmExecPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/wasm_exec.js');
    const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
    eval(wasmExecCode);
    
    const go = new (global as any).Go();
    expect(go).toBeDefined();
    expect(go.importObject).toBeDefined();
  });
  
  test('should load WASM binary', async () => {
    const wasmPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/weewar-cli.wasm');
    const wasmExecPath = path.join(__dirname, STATIC_BASE_PATH + '/wasm/wasm_exec.js');
    
    // Load Go runtime
    const wasmExecCode = fs.readFileSync(wasmExecPath, 'utf8');
    eval(wasmExecCode);
    
    // Load WASM
    const wasmBuffer = fs.readFileSync(wasmPath);
    const arrayBuffer = wasmBuffer.buffer.slice(
      wasmBuffer.byteOffset, 
      wasmBuffer.byteOffset + wasmBuffer.byteLength
    );
    
    const go = new (global as any).Go();
    
    const wasmModule = await WebAssembly.instantiate(arrayBuffer, go.importObject);
    expect(wasmModule).toBeDefined();
    expect(wasmModule.instance).toBeDefined();
    
    console.log('WASM module loaded successfully');
    
    // Try to run it
    const runPromise = go.run(wasmModule.instance);
    
    // Give it some time to initialize
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    console.log('Available global functions:');
    const globalKeys = Object.keys(global).filter(k => k.startsWith('weewar'));
    console.log(globalKeys);
  }, 15000);
});
