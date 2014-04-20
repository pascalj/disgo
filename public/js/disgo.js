(function(){
  var disgo = window.disgo

  function loadDisgo() {
    $each($('[data-disgo-url]'), function(el, i) {
      initializeComments(el)
    });
  }

  function initializeComments(el) {
    var url = el.getAttribute('data-disgo-url')
    $ajax('GET', disgo.base + '/comments?url=' + encodeURIComponent(url), {}, {"Accept": "text/html"}, function(status, result, xhr) {
      if (status != 200) {
        window.console && console.log('Error loading disgo: ' + xhr.status);
        return;
      }
      el.innerHTML += result
      $1('[name=url]', el).setAttribute('value', url)
      $1('form', el).addEventListener('submit', function(e) {
        e.preventDefault()
        submitComment(el)
      })
    });
  }

  function submitComment(el) {
    var form = $1("form", el)
    var data = {
      'name': form.name.value,
      'email': form.email.value,
      'body': form.body.value,
      'url': form.url.value
    }
    $each($('input, textarea', el), function (el, i) { $removeClass(el, 'error') })
    $ajax('POST', disgo.base + '/comments', data, {"Accept": "text/html"}, function(status, result, xhr) {
      if (status != 200) {
        if (disgo.onSubmitError) {
          disgo.onSubmitError(status, result, xhr, form)
        } else {
          var errors = JSON.parse(result);
          for (fieldName in errors['fields']) {
            var field = $1('[name=' + fieldName + ']', el)
            if (field) $addClass(field, 'error')
          }
        }
        return
      }
      form.reset()
      if (disgo.onSubmitSuccess) {
        disgo.onSubmitSuccess(status, result, xhr)
      } else {
        el.innerHTML += result
      }
    })
  }

  function $(sel, ctx) {
    if (ctx) {
      return ctx.querySelectorAll(sel)
    } else {
      return document.querySelectorAll(sel)
    }
  }

  function $1(sel, ctx) {
    return $(sel, ctx)[0]
  }

  function $each(elements, clb) {
    Array.prototype.forEach.call(elements, clb)
  }

  function $addClass(el, className) {
    if (el.classList) {
      el.classList.add(className);
    } else {
      el.className += ' ' + className;
    }
  }

  function $removeClass(el, className) {
    if (el.classList) {
      el.classList.remove(className);
    } else {
      el.className = el.className.replace(new RegExp('(^|\\b)' + className.split(' ').join('|') + '(\\b|$)', 'gi'), ' ');
    }
  }

  function $ajax(method, url, data, headers, handler) {
    var invocation = new XMLHttpRequest();
    var dataString = '';
    for(field in data) {
      dataString += field + '=' + encodeURIComponent(data[field]) + '&'
    }

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
      invocation.send(dataString);
    }
  }

  window[addEventListener ? 'addEventListener' : 'attachEvent'](addEventListener ? 'load' : 'onload', loadDisgo)
})(this);
