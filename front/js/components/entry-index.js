/** @jsx React.DOM */

var React = require('react');
var jQuery = require('jquery');

var utils = require('../lib/utils');

module.exports = React.createClass({
  getInitialState: function () {
    return {
      prev: null,
      next: null,
      today: utils.dateToString(new Date()),
      dates: []
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
    jQuery.ajax({
      type: 'GET',
      url: '/dates/' + date,
      dataType: 'json'
    }).done(function (data) {
      this.setState({
        prev: data.PreviousMonth,
        next: data.NextMonth,
        dates: data.EntryDates
      });
    }.bind(this)).fail(function () {
      console.log('Failed to get entry index.');
    }.bind(this));
  },
  render: function () {
    var items = this.state.dates.map(function (date) {
      var day = utils.extractDay(date.Date);
      if (!date.HasEntry && date.Date !== this.state.today) {
        return <li><span className="mp-date-inactive">{day}</span></li>;
      } else if (date.IsFuture) {
        return <li><span className="mp-date-inactive">{day}</span></li>;
      } else if (date.Date === this.props.date) {
        return <li><span className="mp-date-active">{day}</span></li>;
      } else {
        return <li><a href={"#entries/" + date.Date}>{day}</a></li>;
      }
    }, this);
    if (this.state.prev) {
      items.unshift(
        <li><a href={"#entries/" + this.state.prev}><i className="fa fa-arrow-left"></i></a></li>
      );
    }
    if (this.state.next) {
      items.push(
        <li><a href={"#entries/" + this.state.next}><i className="fa fa-arrow-right"></i></a></li>
      );
    }
    return <ul className="mp-entry-index">{items}</ul>;
  }
});
