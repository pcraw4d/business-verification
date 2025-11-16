const nextJest = require('next/jest')

const createJestConfig = nextJest({
  // Provide the path to your Next.js app to load next.config.js and .env files in your test environment
  dir: './',
})

// Add any custom config to be passed to Jest
const customJestConfig = {
  setupFiles: ['<rootDir>/__tests__/mocks/polyfills.js'], // Load polyfills first
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  // Use jest-fixed-jsdom to restore Node.js globals that jsdom overrides
  // This is needed for MSW to work properly with Node's native fetch
  testEnvironment: 'jest-fixed-jsdom',
  // MSW v2 requires customExportConditions for proper Node.js module resolution
  testEnvironmentOptions: {
    customExportConditions: [''],
  },
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
    // Mock until-async to avoid ES module issues
    '^until-async$': '<rootDir>/__tests__/mocks/until-async.js',
  },
  testMatch: [
    '**/__tests__/**/*.[jt]s?(x)',
    '**/?(*.)+(spec|test).[jt]s?(x)',
  ],
  transformIgnorePatterns: [
    'node_modules/(?!(msw|@mswjs|@mswjs/interceptors|@mswjs/core)/)',
  ],
  collectCoverageFrom: [
    'components/**/*.{js,jsx,ts,tsx}',
    'lib/**/*.{js,jsx,ts,tsx}',
    'hooks/**/*.{js,jsx,ts,tsx}',
    '!**/*.d.ts',
    '!**/node_modules/**',
    '!**/.next/**',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
  },
}

// createJestConfig is exported this way to ensure that next/jest can load the Next.js config which is async
module.exports = createJestConfig(customJestConfig)

