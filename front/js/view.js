/** @jsx React.DOM */

var React = require('react');

var BackboneMixin = require('./backbone-mixin');
var utils = require('./utils');

module.exports = React.createClass({
  mixins: [BackboneMixin],
  getBackboneModels: function () {
    return [this.props.entry];
  },
  render: function () {
    return (
      <div>
        <div dangerouslySetInnerHTML={ { __html: utils.lineBreak(this.props.entry.get('body')) } } />
        <p className="pull-right"><span className="char-count">{this.props.entry.count()}</span> 文字</p>
      </div>
    );
  }
});
