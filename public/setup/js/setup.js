(function() {
  $('#' + $('#database option:selected').val()).addClass('active');
  validateDatabase();

  $('#database').on('change', function() {
    $('#sqlite, #postgresql, #mysql').removeClass('active');
    $('#' + selectedDb()).addClass('active');
    validateDatabase()
  })

  var validateTimer;
  $('form input').on('keyup', function() {
    clearTimeout(validateTimer)
    validateTimer = setTimeout(validateDatabase, 1500);
  })

  function validateDatabase() {
    var data = {};

    switch (selectedDb()) {
      case 'sqlite':
        data.path = $('#sqlite input').val()
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
        console.log(data)
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