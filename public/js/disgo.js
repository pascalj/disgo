
function loadDisgo() {
  $.each($('[data-disgo-url]'), function(i, el) {
    initializeComments(el)
  });

  function initializeComments(el) {
    var url = $(el).attr('data-disgo-url')
    ajax('GET', 'http://localhost:3000/comments', {url: url}, {"Accept": "text/html"}, function(status, result, xhr) {
      if (status != 200) {
        alert('Error ' + xhr.status);
        return;
      }
      el.innerHTML += result
      $('[name=url]', el).attr('value', url)
      $('form', el).on('submit', function(e) {
        e.preventDefault()
        submitComment(el)
      })
    });
  }

  function submitComment(el) {
    var form = $("form", el)
    ajax('POST', 'http://localhost:3000/comments', form.serialize(), {"Accept": "text/html"}, function(status, result, xhr) {
      if (status != 200) {
        var errors = JSON.parse(result);
        for (fieldName in errors['fields']) {
          var field = $('[name=' + fieldName + ']', el)
          if (field) field.addClass('error')
        }
        return
      }
      el.innerHTML += result

      $('form', el).reset()
    })
  }
}

window[addEventListener ? 'addEventListener' : 'attachEvent'](addEventListener ? 'load' : 'onload', loadDisgo)
function ajax(method, url, data, headers, handler) {
  var invocation = new XMLHttpRequest();

  if(invocation) {
    invocation.withCredentials = true;
    invocation.open(method, url, true);
    invocation.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded')
    for (headerName in headers) {
      invocation.setRequestHeader(headerName, headers[headerName])
    }
    invocation.onreadystatechange = function(xhr) {
      if (invocation.readyState == 4) {
        handler(invocation.status, invocation.responseText, invocation);
      }
    }
    invocation.send(data);
  }
}
