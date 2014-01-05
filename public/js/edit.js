/** @jsx React.DOM */

jQuery(function ($) {
  var container = document.getElementById('mp-edit-container');
  if (!container) return;

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
        <div>
          <h2>{this.props.entry.Date}</h2>
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
      <Edit entry={data} />,
      container
    );
  }).fail(function () {
    console.log('Failed to get entry.');
  });

});
