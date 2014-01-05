/** @jsx React.DOM */

jQuery(function ($) {
  var container = document.getElementById('mp-view-container');
  if (!container) return;

  function lineBreak(str) {
    return { __html: str.replace(/\r?\n/g, '<br />') };
  }

  function pad(num) {
    var str = num.toString();
    if (str.length === 1) {
      str = '0' + str;
    }
    return str;
  }

  function dateToString(date) {
    return [date.getFullYear(), date.getMonth() + 1, date.getDate()].map(pad).join('-');
  }

  var View = React.createClass({
    render: function () {
      var today = dateToString(new Date());
      var isEditable = today === this.props.entry.Date;
      var button;
      if (isEditable) {
        button = <a href={"/entries/" + this.props.entry.Date + "/edit"} className="btn btn-default btn-xs mp-edit-button">編集</a>;
      } else {
        button = <button className="btn btn-default btn-xs mp-edit-button" disabled>編集できません</button>;
      }
      return (
        <div>
          <h2>
            {this.props.entry.Date}
            {button}
          </h2>
          <div dangerouslySetInnerHTML={lineBreak(this.props.entry.Body)} />
          <p className="pull-right"><span className="char-count">{this.props.entry.Body.length}</span> 文字</p>
        </div>
      );
    }
  });

  var dateString = window.location.pathname.split('/')[2];
  $.ajax({
    type: 'GET',
    url: '/entries/' + dateString,
    dataType: 'json'
  }).done(function (data) {
    React.renderComponent(
      <View entry={data} />,
      container
    );
  }).fail(function () {
    console.log('Failed to get entry.');
  });
});
