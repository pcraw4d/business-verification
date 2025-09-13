/**
 * Webpack Configuration for KYB Platform Frontend
 * Optimized for code splitting, tree shaking, and performance
 */

const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CompressionPlugin = require('compression-webpack-plugin');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = (env, argv) => {
    const isProduction = argv.mode === 'production';
    const isDevelopment = !isProduction;

    return {
        mode: isProduction ? 'production' : 'development',
        
        entry: {
            // Core application entry points
            'merchant-portfolio': './components/merchant-portfolio.js',
            'merchant-dashboard': './components/merchant-dashboard.js',
            'merchant-comparison': './components/merchant-comparison.js',
            'bulk-operations': './components/bulk-operations.js',
            
            // Utility modules
            'lazy-loader': './components/lazy-loader.js',
            'virtual-scroller': './components/virtual-scroller.js',
            'bundle-optimizer': './components/bundle-optimizer.js',
            
            // Shared components
            'merchant-search': './components/merchant-search.js',
            'portfolio-type-filter': './components/portfolio-type-filter.js',
            'risk-level-indicator': './components/risk-level-indicator.js',
            'session-manager': './components/session-manager.js',
            'coming-soon-banner': './components/coming-soon-banner.js',
            'mock-data-warning': './components/mock-data-warning.js',
            
            // Main application bundles
            'app': './index.js',
            'dashboard-hub': './dashboard-hub.js'
        },

        output: {
            path: path.resolve(__dirname, 'dist'),
            filename: isProduction ? '[name].[contenthash:8].js' : '[name].js',
            chunkFilename: isProduction ? '[name].[contenthash:8].chunk.js' : '[name].chunk.js',
            publicPath: '/',
            clean: true,
            library: {
                type: 'umd',
                name: 'KYBPlatform'
            }
        },

        optimization: {
            minimize: isProduction,
            minimizer: [
                new TerserPlugin({
                    terserOptions: {
                        compress: {
                            drop_console: isProduction,
                            drop_debugger: isProduction,
                            pure_funcs: isProduction ? ['console.log', 'console.info'] : []
                        },
                        mangle: {
                            safari10: true
                        },
                        format: {
                            comments: false
                        }
                    },
                    extractComments: false
                })
            ],

            // Code splitting configuration
            splitChunks: {
                chunks: 'all',
                cacheGroups: {
                    // Vendor libraries
                    vendor: {
                        test: /[\\/]node_modules[\\/]/,
                        name: 'vendors',
                        chunks: 'all',
                        priority: 10,
                        reuseExistingChunk: true
                    },
                    
                    // Common modules
                    common: {
                        name: 'common',
                        minChunks: 2,
                        chunks: 'all',
                        priority: 5,
                        reuseExistingChunk: true
                    },
                    
                    // Merchant-specific modules
                    merchant: {
                        test: /[\\/]components[\\/]merchant/,
                        name: 'merchant',
                        chunks: 'all',
                        priority: 8,
                        reuseExistingChunk: true
                    },
                    
                    // Utility modules
                    utils: {
                        test: /[\\/]components[\\/](lazy-loader|virtual-scroller|bundle-optimizer)/,
                        name: 'utils',
                        chunks: 'all',
                        priority: 7,
                        reuseExistingChunk: true
                    },
                    
                    // Default chunk
                    default: {
                        minChunks: 2,
                        priority: -20,
                        reuseExistingChunk: true
                    }
                }
            },

            // Runtime chunk for better caching
            runtimeChunk: {
                name: 'runtime'
            },

            // Module concatenation for better tree shaking
            concatenateModules: isProduction,

            // Side effects optimization
            sideEffects: false
        },

        module: {
            rules: [
                // JavaScript/TypeScript files
                {
                    test: /\.(js|ts)$/,
                    exclude: /node_modules/,
                    use: {
                        loader: 'babel-loader',
                        options: {
                            presets: [
                                ['@babel/preset-env', {
                                    targets: {
                                        browsers: ['> 1%', 'last 2 versions', 'not ie <= 8']
                                    },
                                    modules: false, // Let webpack handle modules
                                    useBuiltIns: 'usage',
                                    corejs: 3
                                }]
                            ],
                            plugins: [
                                '@babel/plugin-syntax-dynamic-import',
                                '@babel/plugin-proposal-class-properties',
                                '@babel/plugin-proposal-optional-chaining',
                                '@babel/plugin-proposal-nullish-coalescing-operator'
                            ]
                        }
                    }
                },

                // CSS files
                {
                    test: /\.css$/,
                    use: [
                        isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
                        {
                            loader: 'css-loader',
                            options: {
                                modules: false,
                                sourceMap: isDevelopment
                            }
                        },
                        {
                            loader: 'postcss-loader',
                            options: {
                                postcssOptions: {
                                    plugins: [
                                        'autoprefixer',
                                        'cssnano'
                                    ]
                                }
                            }
                        }
                    ]
                },

                // SCSS files
                {
                    test: /\.scss$/,
                    use: [
                        isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
                        {
                            loader: 'css-loader',
                            options: {
                                modules: false,
                                sourceMap: isDevelopment
                            }
                        },
                        {
                            loader: 'sass-loader',
                            options: {
                                sourceMap: isDevelopment
                            }
                        }
                    ]
                },

                // Images
                {
                    test: /\.(png|jpe?g|gif|svg|webp)$/i,
                    type: 'asset',
                    parser: {
                        dataUrlCondition: {
                            maxSize: 8 * 1024 // 8KB
                        }
                    },
                    generator: {
                        filename: isProduction ? 'images/[name].[contenthash:8][ext]' : 'images/[name][ext]'
                    }
                },

                // Fonts
                {
                    test: /\.(woff|woff2|eot|ttf|otf)$/i,
                    type: 'asset/resource',
                    generator: {
                        filename: isProduction ? 'fonts/[name].[contenthash:8][ext]' : 'fonts/[name][ext]'
                    }
                },

                // HTML files
                {
                    test: /\.html$/,
                    use: {
                        loader: 'html-loader',
                        options: {
                            minimize: isProduction,
                            removeComments: isProduction,
                            collapseWhitespace: isProduction
                        }
                    }
                }
            ]
        },

        plugins: [
            // Extract CSS to separate files
            new MiniCssExtractPlugin({
                filename: isProduction ? '[name].[contenthash:8].css' : '[name].css',
                chunkFilename: isProduction ? '[name].[contenthash:8].chunk.css' : '[name].chunk.css'
            }),

            // Gzip compression
            ...(isProduction ? [
                new CompressionPlugin({
                    test: /\.(js|css|html|svg)$/,
                    algorithm: 'gzip',
                    threshold: 1024,
                    minRatio: 0.8
                })
            ] : []),

            // Bundle analyzer (only in development or when explicitly requested)
            ...(process.env.ANALYZE ? [
                new BundleAnalyzerPlugin({
                    analyzerMode: 'static',
                    openAnalyzer: false,
                    reportFilename: 'bundle-report.html'
                })
            ] : [])
        ],

        resolve: {
            extensions: ['.js', '.ts', '.json'],
            alias: {
                '@': path.resolve(__dirname, 'components'),
                '@utils': path.resolve(__dirname, 'components/utils'),
                '@merchant': path.resolve(__dirname, 'components/merchant'),
                '@styles': path.resolve(__dirname, 'styles')
            }
        },

        devtool: isDevelopment ? 'eval-source-map' : 'source-map',

        devServer: {
            static: {
                directory: path.join(__dirname, 'dist')
            },
            compress: true,
            port: 3000,
            hot: true,
            historyApiFallback: true,
            headers: {
                'Cache-Control': 'no-cache'
            }
        },

        performance: {
            hints: isProduction ? 'warning' : false,
            maxEntrypointSize: 512000, // 500KB
            maxAssetSize: 512000, // 500KB
            assetFilter: (assetFilename) => {
                return !assetFilename.endsWith('.map');
            }
        },

        stats: {
            colors: true,
            modules: false,
            children: false,
            chunks: false,
            chunkModules: false,
            entrypoints: false,
            assets: true,
            version: false,
            hash: false,
            timings: true,
            builtAt: false
        }
    };
};
