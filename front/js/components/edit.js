/** @jsx React.DOM */

var React = require('react');

var BackboneMixin = require('../lib/backbone-mixin');
var utils = require('../lib/utils');

module.exports = React.createClass({
  mixins:[BackboneMixin],
  getInitialState: function () {
    return { dirty: undefined };
  },
  getBackboneModels: function () {
    window.a = this.props.entry;
    return [this.props.entry];
  },
  componentDidMount: function () {
    this.wait();
  },
  // auto: boolean to indicate whether it's auto save or not.
  save: function (auto) {
    if (!this.state.dirty) {
      console.log('nothing to save');
      if (auto) {
        this.wait();
      } else {
        window.location = '#/entries/' + this.props.entry.get('date');
      }
      return;
    }

    this.setState({ dirty: false });

    // TODO: Block editing if not auto save.
    this.props.entry.save().done(function () {
      console.log('save success');
      if (!auto)  window.location = '#/entries/' + this.props.entry.get('date');
    }.bind(this)).fail(function () {
      console.log('save failure', arguments[0].responseText);
      this.setState({ dirty: true });
    }.bind(this)).always(function () {
      if (auto && this.isMounted()) this.wait();
    }.bind(this));
  },
  wait: function () {
    setTimeout(this.save.bind(this, true), 15 * 1000);
  },
  handleChange: function (e) {
    this.props.entry.set('body', e.target.value);
    this.setState({ dirty: true });
  },
  render: function () {
    var status;
    if (this.state.dirty === undefined) {
      status = '';
    } else if (this.state.dirty) {
      status = <span><i className="fa fa-pencil" /> 未保存</span>;
    } else {
      status = <span><i className="fa fa-check" /> 保存済</span>;
    }
    return (
      <div className="form" id="mp-entry-form">
        <div className="form-group" id="mp-entry-body">
          <textarea name="body" rows="20" cols="80" className="form-control"
                    value={this.props.entry.get('body')} onChange={this.handleChange} />
        </div>
        <div className="form-group">
          <button className="btn btn-default" onClick={this.save.bind(this, false)}>完了</button>
          <span id="mp-entry-status">{status}</span>
          <p className="pull-right">
            <span id="mp-char-count">{this.props.entry.count()}</span> 文字
          </p>
        </div>
      </div>
    );
  }
});