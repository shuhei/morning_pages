jQuery(function ($) {
  var isDirty = false;
  var $entryForm = $('#mp-entry-form');
  var $entryBody = $('#mp-entry-body').on('keyup', function () {
    becomeDirty();
    updateCount();
  });
  var $textarea = $entryBody.find('textarea');
  var $charCount = $('#mp-char-count');
  var $status = $('#mp-entry-status');

  function updateCount() {
    var count = $textarea.val().length;
    $charCount.text(count);
  }

  function becomeDirty() {
    if (isDirty) return;
    isDirty = true;
    $status.html('<i class="fa fa-pencil"></i> 未保存')
  }

  function becomeClean() {
    if (!isDirty) return;
    isDirty = false;
    $status.html('<i class="fa fa-check"></i> 保存済')
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
    setTimeout(autoSave, 15 * 1000);
  }

  becomeClean();
  wait();
});
