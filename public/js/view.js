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
      return (
        <div>
          <div dangerouslySetInnerHTML={lineBreak(this.props.entry.Body)} />
          <p className="pull-right"><span className="char-count">{this.props.entry.Body.length}</span> 文字</p>
        </div>
      );
    }
  });

  var Edit = React.createClass({
    getInitialState: function () {
      return {
        body: this.props.entry.Body,
        dirty: undefined
      };
    },
    componentDidMount: function () {
      this.wait();
    },
    autoSave: function () {
      if (!this.state.dirty) {
        console.log('nothing to post');
        this.wait();
        return;
      }

      this.setState({ dirty: false });

      $.ajax({
        type: 'POST',
        url: "/entries/" + this.props.entry.Date,
        data: { body: this.state.body }
      }).done(function () {
        console.log('success');
      }.bind(this)).fail(function () {
        console.log('failure');
        this.setState({ dirty: true });
      }.bind(this)).always(function () {
        if (this.isMounted()) this.wait();
      }.bind(this));
    },
    wait: function () {
      setTimeout(this.autoSave.bind(this), 15 * 1000);
    },
    onKeyUp: function (e) {
      this.setState({ body: e.target.value, dirty: true })
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
        <form action={"/entries/" + this.props.entry.Date} method="POST" className="form" id="mp-entry-form">
          <div className="form-group" id="mp-entry-body">
            <textarea name="body" rows="20" cols="80" className="form-control" onKeyUp={this.onKeyUp}>{this.state.body}</textarea>
          </div>
          <div className="form-group">
            <input type="submit" value="完了" className="btn btn-default" />
            <span id="mp-entry-status">{status}</span>
            <p className="pull-right">
              <span id="mp-char-count">{this.state.body.length}</span> 文字
            </p>
          </div>
        </form>
      );
    }
  });

  var EntryApp = React.createClass({
    getInitialState: function () {
      return { editing: false };
    },
    edit: function () {
      this.setState({ editing: true });
    },
    render: function () {
      var today = dateToString(new Date());
      var isEditable = today === this.props.entry.Date;
      var button;
      if (this.state.editing) {
        button = '';
      } else if (isEditable) {
        button = <button className="btn btn-default btn-xs mp-edit-button" onClick={this.edit}>編集</button>;
      } else {
        button = <button className="btn btn-default btn-xs mp-edit-button" disabled>編集できません</button>;
      }

      return (
        <div>
          <h2>{this.props.entry.Date} {button}</h2>
          {this.state.editing ? <Edit entry={this.props.entry} /> : <View entry={this.props.entry} />}
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
      <EntryApp entry={data} />,
      container
    );
  }).fail(function () {
    console.log('Failed to get entry.');
  });
});
