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

  function extractDay(dateString) {
    return dateString.split('-')[2].replace(/^0/, '');
  }

  var EntryIndex = React.createClass({
    getInitialState: function () {
      return {
        prev: null,
        next: null,
        today: dateToString(new Date()),
        dates: []
      };
    },
    componentDidMount: function () {
      $.ajax({
        type: 'GET',
        url: '/dates/' + this.props.date,
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
        var day = extractDay(date.Date);
        if (!date.HasEntry && date.Date !== this.state.today) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.IsFuture) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.Date === this.props.date) {
          return <li><span className="mp-date-active">{day}</span></li>;
        } else {
          return <li><a href={"/entries/" + date.Date}>{day}</a></li>;
        }
      }, this);
      if (this.state.prev) {
        items.unshift(
          <li><a href={"/entries/" + this.state.prev}><i className="fa fa-arrow-left"></i></a></li>
        );
      }
      if (this.state.next) {
        items.push(
          <li><a href={"/entries/" + this.state.next}><i className="fa fa-arrow-right"></i></a></li>
        );
      }
      return <ul className="mp-entry-index">{items}</ul>;
    }
  });

  var Entry = Backbone.Model.extend({
    idAttribute: 'Date',
    urlRoot: '/entries',
    defaults: {
      Body: ''
    },
    count: function () {
      return this.get('Body').length;
    }
  });

  // From React TodoMVC Backbone example.
  // https://github.com/facebook/react/tree/master/examples/todomvc-backbone
  //
  // An example generic Mixin that you can add to any component that should react
  // to changes in a Backbone component. The use cases we've identified thus far
  // are for Collections -- since they trigger a change event whenever any of
  // their constituent items are changed there's no need to reconcile for regular
  // models. One caveat: this relies on getBackboneModels() to always return the
  // same model instances throughout the lifecycle of the component. If you're
  // using this mixin correctly (it should be near the top of your component
  // hierarchy) this should not be an issue.
  var BackboneMixin = {
    componentDidMount: function() {
      // Whenever there may be a change in the Backbone data, trigger a reconcile.
      this.getBackboneModels().forEach(function(model) {
        model.on('add change remove', this.forceUpdate.bind(this, null), this);
      }, this);
    },

    componentWillUnmount: function() {
      // Ensure that we clean up any dangling references when the component is
      // destroyed.
      this.getBackboneModels().forEach(function(model) {
        model.off(null, null, this);
      }, this);
    }
  };

  var View = React.createClass({
    mixins: [BackboneMixin],
    getBackboneModels: function () {
      return [this.props.entry];
    },
    render: function () {
      return (
        <div>
          <div dangerouslySetInnerHTML={lineBreak(this.props.entry.get('Body'))} />
          <p className="pull-right"><span className="char-count">{this.props.entry.count()}</span> 文字</p>
        </div>
      );
    }
  });

  var Edit = React.createClass({
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
    autoSave: function () {
      if (!this.state.dirty) {
        console.log('nothing to autosave');
        this.wait();
        return;
      }

      this.setState({ dirty: false });

      this.props.entry.save().done(function () {
        console.log('autosave success');
      }.bind(this)).fail(function () {
        console.log('autosave failure', arguments[0].responseText);
        this.setState({ dirty: true });
      }.bind(this)).always(function () {
        if (this.isMounted()) this.wait();
      }.bind(this));
    },
    save: function () {
      // TODO: Block editing.
      this.props.entry.save().done(function () {
        console.log('save success');
        this.setState({ dirty: false });
        window.location = '#/';
      }.bind(this)).fail(function () {
        console.log('save failure', arguments[0].responseText);
        this.setState({ dirty: true });
      }.bind(this));
    },
    wait: function () {
      setTimeout(this.autoSave.bind(this), 15 * 1000);
    },
    onKeyUp: function (e) {
      this.props.entry.set('Body', e.target.value);
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
            <textarea name="body" rows="20" cols="80" className="form-control" onKeyUp={this.onKeyUp}>
              {this.props.entry.get('Body')}
            </textarea>
          </div>
          <div className="form-group">
            <button className="btn btn-default" onClick={this.save}>完了</button>
            <span id="mp-entry-status">{status}</span>
            <p className="pull-right">
              <span id="mp-char-count">{this.props.entry.count()}</span> 文字
            </p>
          </div>
        </div>
      );
    }
  });

  var EntryApp = React.createClass({
    getInitialState: function () {
      return { editing: false };
    },
    componentDidMount: function () {
      this.router = new Router({
        '/': this.view.bind(this),
        '/edit': this.edit.bind(this)
      });
      this.router.init();
    },
    componentWillUnmount: function () {
      this.router.destroy();
      delete this.router;
    },
    view: function () {
      this.setState({ editing: false });
    },
    edit: function () {
      this.setState({ editing: true });
    },
    render: function () {
      var today = dateToString(new Date());
      var isEditable = today === this.props.entry.get('Date');
      var button;
      if (this.state.editing) {
        button = '';
      } else if (isEditable) {
        button = <a className="btn btn-default btn-xs mp-edit-button" href="#/edit">編集</a>;
      } else {
        button = <button className="btn btn-default btn-xs mp-edit-button" disabled>編集できません</button>;
      }

      return (
        <div>
          <EntryIndex date={this.props.entry.get('Date')} />
          <h2>{this.props.entry.get('Date')} {button}</h2>
          {this.state.editing ? 
            <Edit entry={this.props.entry} ref="edit" /> :
            <View entry={this.props.entry} ref="view" />
          }
        </div>
      );
    }
  });

  var dateString = window.location.pathname.split('/')[2];
  var entry = new Entry({ Date: dateString });
  entry.fetch().done(function () {
    React.renderComponent(
      <EntryApp entry={entry} />,
      container
    );
  }).fail(function () {
    console.log('Failed to get entry.');
  });
});
