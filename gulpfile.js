// @ts-check

const gulp = require('gulp')

const { graphQlOperations, schema, watchGraphQlOperations } = require('./client/shared/gulpfile')
const { build, developmentServer, generate, watchGenerators } = require('./client/web/gulpfile')

/**
 * Generates files needed for builds whenever files change.
 */
const watchGenerate = gulp.series(generate, watchGenerators)

/**
 * Watches everything and rebuilds on file changes.
 */
const development = gulp.series(generate, gulp.parallel(watchGenerators, developmentServer))

module.exports = {
  generate,
  watchGenerate,
  build,
  dev: development,
  schema,
  graphQlOperations,
  watchGraphQlOperations,
}
