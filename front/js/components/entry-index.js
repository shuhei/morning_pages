/** @jsx React.DOM */

var React = require('react');

var EntryList = require('../models/entry-list');
var utils = require('../lib/utils');

module.exports = React.createClass({
  getInitialState: function () {
    return {
      from: null,
      to: null,
      today: utils.dateToString(new Date()),
      entries: new EntryList()
    };
  },
  componentDidMount: function () {
    this.fetchDates(this.props.date);
  },
  componentWillReceiveProps: function (nextProps) {
    if (this.props.date !== nextProps.date) {
      this.fetchDates(nextProps.date);
    }
  },
  fetchDates: function (date) {
    var d = utils.parseDate(date);
    var from = utils.beginningOfMonth(d);
    var to = utils.endOfMonth(d);
    if (this.state.from && this.state.to && this.state.from <= d && d <= this.state.to) {
      return;
    }
    var query = { from: utils.dateToString(from), to: utils.dateToString(to) };
    var entries = new EntryList([], query);
    entries.fetch().done(function () {
      this.setState({
        from: from,
        to: to,
        entries: entries
      });
    }.bind(this)).fail(function () {
      console.log('Failed to get entry index.');
    }.bind(this));
  },
  render: function () {
    if (!this.state.from || !this.state.to) {
      return <div />;
    }

    var items = [];
    for (var day = 1, l = this.state.to.getDate(); day <= l; day++) {
      var date = utils.dateString(this.state.from.getFullYear(), this.state.from.getMonth(), day);
      if (this.state.today < date) {
        break;
      }
      var entry = this.state.entries.findWhere({ date: date });
      if (date === this.props.date) {
        items.push(<li key={date}><span className="mp-date-active">{day}</span></li>);
      } else if (entry || date === this.state.today) {
        items.push(<li key={date}><a href={"#entries/" + date}>{day}</a></li>);
      } else {
        items.push(<li key={date}><span className="mp-date-inactive">{day}</span></li>);
      }
    }

    var prev = utils.dateToString(utils.beginningOfMonth(utils.prevDate(this.state.from)));
    items.unshift(
      <li key={prev}><a href={"#entries/" + prev}><i className="fa fa-arrow-left"></i></a></li>
    );

    var next = utils.dateToString(utils.nextDate(this.state.to));
    if (next <= this.state.today) {
      items.push(
        <li key={next}><a href={"#entries/" + next}><i className="fa fa-arrow-right"></i></a></li>
      );
    }
    return <ul className="mp-entry-index">{items}</ul>;
  }
});
