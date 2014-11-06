(function() {
  $('#' + $('#database option:selected').val()).addClass('active');
  validateDatabase();

  $('#database').on('change', function() {
    $('#sqlite3, #postgresql, #mysql').removeClass('active');
    $('#' + selectedDb()).addClass('active');
    validateDatabase();
  })

  $('#finish').on('click', function(e) {
    e.preventDefault();
    setDatabase();
  })

  var validateTimer;
  $('form input').on('keyup', function() {
    clearTimeout(validateTimer)
    validateTimer = setTimeout(validateDatabase, 1500);
  })

  function validateDatabase() {
    var data = {};

    switch (selectedDb()) {
      case 'sqlite3':
        data.path = $('#sqlite3 input').val();
        break;
      case 'postgresql':
        data = getData($('#postgresql'));
        break;
      case 'mysql':
        data = getData($('#mysql'));
        break;
    }

    data.driver = selectedDb();

    $.ajax({
      url: '/setup/database',
      data: data,
      success: function(data) {
        if (data == true) {
          $('#finish').addClass('success').attr('disabled', false);
          $('.wrong-data').hide();
        } else {
          $('#finish').removeClass('success').attr('disabled', true)
          $('.wrong-data').show();
        }
      },
    });
  }

  function setDatabase() {
    $('#finish').attr('disabled', true);
    var data = {};

    switch (selectedDb()) {
      case 'sqlite3':
        data.path = $('#sqlite3 input').val()
        break;
      case 'postgresql':
        data = getData($('#postgresql'));
        break;
      case 'mysql':
        data = getData($('#mysql'));
        break;
    }

    data.driver = selectedDb();

    $.ajax({
      url: '/setup/database',
      data: data,
      type: 'POST',
      success: function(data) {
        if (data == true) {
          setTimeout('window.location.href=window.location.href', 500)
        } else {
          $('#finish').attr('disabled', false);
        }
      },
    });
  }

  function selectedDb() {
    return $('#database option:selected').val();
  }

  function getData(parent) {
    return {
      host: $('.host', parent).val(),
      port: $('.port', parent).val(),
      username: $('.username', parent).val(),
      password: $('.password', parent).val(),
      database: $('.database', parent).val()
    }
  }
})()
