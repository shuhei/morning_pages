/** @jsx React.DOM */

jQuery(function ($) {
  var container = document.getElementById('mp-view-container');
  if (!container) return;

  function lineBreak(str) {
    return str.replace(/\r?\n/g, '<br />');
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
      this.fetchDates(this.props.date);
    },
    componentWillReceiveProps: function (nextProps) {
      if (this.props.date !== nextProps.date) {
        this.fetchDates(nextProps.date);
      }
    },
    fetchDates: function (date) {
      $.ajax({
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
        var day = extractDay(date.Date);
        if (!date.HasEntry && date.Date !== this.state.today) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.IsFuture) {
          return <li><span className="mp-date-inactive">{day}</span></li>;
        } else if (date.Date === this.props.date) {
          return <li><span className="mp-date-active">{day}</span></li>;
        } else {
          return <li><a href={"#/" + date.Date}>{day}</a></li>;
        }
      }, this);
      if (this.state.prev) {
        items.unshift(
          <li><a href={"#/" + this.state.prev}><i className="fa fa-arrow-left"></i></a></li>
        );
      }
      if (this.state.next) {
        items.push(
          <li><a href={"#/" + this.state.next}><i className="fa fa-arrow-right"></i></a></li>
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
          <div dangerouslySetInnerHTML={ { __html: lineBreak(this.props.entry.get('Body')) } } />
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
    // auto: boolean to indicate whether it's auto save or not.
    save: function (auto) {
      if (!this.state.dirty) {
        console.log('nothing to save');
        if (auto) {
          this.wait();
        } else {
          window.location = '#/' + this.props.entry.get('Date');
        }
        return;
      }

      this.setState({ dirty: false });

      // TODO: Block editing if not auto save.
      this.props.entry.save().done(function () {
        console.log('save success');
        if (!auto)  window.location = '#/' + this.props.entry.get('Date');
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
            <textarea name="body" rows="20" cols="80" className="form-control"
                      value={this.props.entry.get('Body')} onChange={this.handleChange} />
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

  var EntryApp = React.createClass({
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
      // TODO: Avoid fetching multiple entries at the same time.
      var entry = new Entry({ Date: date });
      entry.fetch().done(function () {
        this.setState({ entry: entry });
      }.bind(this)).fail(function () {
        console.log('Failed to get entry.');
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

      var today = dateToString(new Date());
      var isEditable = today === this.state.entry.get('Date');
      var button;
      if (this.state.editing) {
        button = '';
      } else if (isEditable) {
        var editPath = '#/' + this.state.entry.get('Date') + '/edit';
        button = <a className="btn btn-default btn-xs mp-edit-button" href={editPath}>編集</a>;
      } else {
        button = <button className="btn btn-default btn-xs mp-edit-button" disabled>編集できません</button>;
      }
      return (
        <div>
          <EntryIndex date={this.state.date} />
          <h2>{this.state.entry.get('Date')} {button}</h2>
          {this.state.editing ? 
            <Edit entry={this.state.entry} /> :
            <View entry={this.state.entry} />
          }
        </div>
      );
    }
  });

  var app = React.renderComponent(
    <EntryApp />,
    container
  );

  function showToday() {
    window.location = '#/' + dateToString(new Date());
  }

  var router = new Router({
    '/': showToday,
    '/:date': app.show.bind(app),
    '/:date/edit': app.edit.bind(app),
    // Facebook callback redirects to URL with '#_=_'.
    // http://stackoverflow.com/questions/7131909/facebook-callback-appends-to-return-url
    '_=_': showToday
  });
  router.init('/');
});
