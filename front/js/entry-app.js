/** @jsx React.DOM */

var React = require('react');

var Entry = require('./entry');
var EntryIndex = require('./entry-index');
var View = require('./view');
var Edit = require('./edit');
var utils = require('./utils');

module.exports = React.createClass({
  getInitialState: function () {
    return {
      editing: false,
      date: undefined,
      entry: undefined
    };
  },
  componentWillUpdate: function (nextProps, nextState) {
    if (this.state.date !== nextState.date) {
      console.log('date changed from', this.state.date, 'to', nextState.date);
      this.fetchEntry(nextState.date);
    }
  },
  fetchEntry: function (date) {
    // TODO: Fetch enrties on ahead.
    var entry = new Entry({ date: date });
    entry.fetch().done(function () {
      this.setState({ entry: entry });
    }.bind(this)).fail(function (xhr) {
      if (xhr.status === 404) {
        this.setState({ entry: entry });
      } else {
        console.log('Failed to get entry.');
      }
    }.bind(this));
  },
  show: function (date) {
    this.setState({ date: date, editing: false });
  },
  edit: function (date) {
    this.setState({ date: date, editing: true });
  },
  render: function () {
    if (!this.state.entry) {
      return <div></div>;
    }

    var today = utils.dateToString(new Date());
    var isEditable = today === this.state.entry.get('date');
    var button;
    if (this.state.editing) {
      button = '';
    } else if (isEditable) {
      var editPath = '#entries/' + this.state.entry.get('date') + '/edit';
      button = <a className="btn btn-default btn-xs mp-edit-button" href={editPath}>編集</a>;
    } else {
      button = <button className="btn btn-default btn-xs mp-edit-button" disabled>編集できません</button>;
    }
    return (
      <div>
        <EntryIndex date={this.state.date} />
        <h2>{this.state.entry.get('date')} {button}</h2>
        {this.state.editing ? 
          <Edit entry={this.state.entry} /> :
          <View entry={this.state.entry} />
        }
      </div>
    );
  }
});
