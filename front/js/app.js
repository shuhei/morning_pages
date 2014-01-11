/** @jsx React.DOM */

var jQuery = require('jquery')(window);
var Backbone = require('./lib/backbone-shim');
var React = require('react');

var EntryApp = require('./components/entry-app');
var utils = require('./lib/utils');

jQuery(function ($) {
  var container = document.getElementById('mp-view-container');
  if (!container) return;

  var app = React.renderComponent(
    <EntryApp />,
    container
  );

  var AppRouter = Backbone.Router.extend({
    initialize: function (options) {
      this.app = options.app;
    },
    routes: {
      '': 'showToday',
      'entries/:date': 'show',
      'entries/:date/edit': 'edit',
      // Facebook callback redirects to URL with '#_=_'.
      // http://stackoverflow.com/questions/7131909/facebook-callback-appends-to-return-url
      '_=_': 'showToday'
    },
    show: function (date) {
      this.app.show(date);
    },
    edit: function (date) {
      this.app.edit(date);
    },
    showToday: function () {
      this.navigate('entries/' + utils.dateToString(new Date()), { trigger: true });
    }
  });
  new AppRouter({ app: app });
  Backbone.history.start();
});
