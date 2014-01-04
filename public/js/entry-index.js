/** @jsx React.DOM */

jQuery(function ($) {
  var container = document.getElementById('mp-entry-index-container');
  if (!container) return;

  function extractDay(dateString) {
    return dateString.split('-')[2].replace(/^0/, '');
  }

  var EntryIndex = React.createClass({
    render: function () {
      var items = this.props.dates.map(function (date) {
        var day = extractDay(date.Date);
        if (!date.HasEntry && date.Date !== this.props.today) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.IsFuture) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.Date === this.props.date) {
          return <li><span className="mp-date-active">{day}</span></li>;
        } else {
          return <li><a href={"/entries/" + date.Date}>{day}</a></li>;
        }
      }, this);
      if (this.props.prev) {
        items.unshift(
          <li><a href={"/entries/" + this.props.prev}><i className="fa fa-arrow-left"></i></a></li>
        );
      }
      if (this.props.next) {
        items.push(
          <li><a href={"/entries/" + this.props.next}><i className="fa fa-arrow-right"></i></a></li>
        );
      }
      return <ul className="mp-entry-index">{items}</ul>;
    }
  });

  var dateString = window.location.pathname.split('/')[2];
  $.ajax({
    type: 'GET',
    url: '/dates/' + dateString,
    dataType: 'json'
  }).done(function (data) {
    React.renderComponent(
      <EntryIndex prev={data.PreviousMonth}
                  next={data.NextMonth}
                  today={data.Today}
                  date={dateString}
                  dates={data.EntryDates} />,
      container
    );
  }).fail(function () {
    console.log('Failed to show entry index.');
  });
});
