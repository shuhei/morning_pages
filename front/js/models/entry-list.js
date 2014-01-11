var _ = require('underscore');
var Backbone = require('../lib/backbone-shim');
var Entry = require('./entry');

module.exports = Backbone.Collection.extend({
  model: Entry,
  initialize: function (models, options) {
    this.from = options.from;
    this.to = options.to;
  },
  url: function () {
    var params = { from: this.from, to: this.to };
    var query = _.map(params, function (v, k) {
      return k + '=' + encodeURIComponent(v);
    }).join('&');
    return '/entries?' + query;
  }
});
