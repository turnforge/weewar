module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  roots: ['<rootDir>/frontend', '<rootDir>/tests'],
  testMatch: [
    '**/__tests__/**/*.ts',
    '**/?(*.)+(spec|test).ts'
  ],
  transform: {
    '^.+\\.ts$': 'ts-jest',
  },
  collectCoverageFrom: [
    'frontend/components/**/*.ts',
    '!frontend/components/**/*.d.ts',
    '!frontend/components/ComponentIsolationTest.ts',
    '!**/node_modules/**'
  ],
  coverageDirectory: 'coverage',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  // Increase timeout for WASM-related tests
  testTimeout: 10000,
  // Allow tests to run in parallel but limit concurrency for WASM tests
  maxWorkers: 2,
  // Ensure proper cleanup between tests
  clearMocks: true,
  restoreMocks: true
};