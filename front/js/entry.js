var Backbone = require('./backbone-shim');

module.exports = Backbone.Model.extend({
  url: function () {
    return '/entries/' + this.get('date');
  },
  defaults: {
    body: ''
  },
  count: function () {
    return this.get('body').length;
  }
});
