const { getDefaultConfig } = require('expo/metro-config')
const path = require('path')

const projectRoot = __dirname
const monorepoRoot = path.resolve(projectRoot, '..')

const config = getDefaultConfig(projectRoot)

// Watch monorepo packages
config.watchFolders = [monorepoRoot]

// Ensure Metro resolves from mobile's node_modules first (prevents duplicate React)
config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(monorepoRoot, 'node_modules'),
]

// Prevent duplicate packages — force resolution from mobile's node_modules
config.resolver.resolveRequest = (context, moduleName, platform) => {
  if (moduleName === 'react' || moduleName === 'react-native' || moduleName.startsWith('@tanstack/react-query') || moduleName === 'zustand') {
    return context.resolveRequest(
      { ...context, originModulePath: path.resolve(projectRoot, 'index.js') },
      moduleName,
      platform,
    )
  }
  return context.resolveRequest(context, moduleName, platform)
}

module.exports = config
