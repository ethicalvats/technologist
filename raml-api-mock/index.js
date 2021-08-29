const express = require('express')
const osprey = require('osprey')
const join = require('path').join

const PORT = process.env.PORT || 3000

const router = osprey.Router()

osprey.loadFile(join(__dirname, 'api-v1.raml'))
  .then(function (middleware) {
    const app = express()

    app.use('/', middleware, router)

    app.listen(PORT, function () {
      console.log('Application listening on ' + PORT + '...')
    })
  })