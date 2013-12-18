jQuery(function ($) {
  var isDirty = false;
  var $entryForm = $('#mp-entry-form');
  var $entryBody = $('#mp-entry-body').on('keyup', function () {
    becomeDirty();
    updateCount();
  });
  var $textarea = $entryBody.find('textarea');
  var $charCount = $entryBody.find('.char-count');
  var $saveButton = $('.mp-save-button');

  function updateCount() {
    var count = $textarea.val().length;
    $charCount.text(count);
  }

  function becomeDirty() {
    isDirty = true;
    $saveButton.removeClass('disabled');
  }

  function becomeClean() {
    isDirty = false;
    $saveButton.addClass('disabled');
  }

  function autoSave() {
    if (!isDirty) {
      console.log('nothing to post');
      wait();
      return;
    }

    var options = {
      type: 'POST',
      url: $entryForm.attr('action'),
      data: { body: $textarea.val() }
    };

    becomeClean();

    $.ajax(options).done(function () {
      console.log('success');
    }).fail(function () {
      console.log('failure');
      becomeDirty();
    }).always(function () {
      wait();
    });
  }

  function wait() {
    setTimeout(autoSave, 5000);
  }

  becomeClean();
  wait();
});
