/*
 *  Copyright 2012-2013 (c) Pierre Duquesne <stackp@online.fr>
 *  Licensed under the New BSD License.
 *  https://github.com/stackp/promisejs
 */
(function(a){function b(){this._callbacks=[];}b.prototype.then=function(a,c){var d;if(this._isdone)d=a.apply(c,this.result);else{d=new b();this._callbacks.push(function(){var b=a.apply(c,arguments);if(b&&typeof b.then==='function')b.then(d.done,d);});}return d;};b.prototype.done=function(){this.result=arguments;this._isdone=true;for(var a=0;a<this._callbacks.length;a++)this._callbacks[a].apply(null,arguments);this._callbacks=[];};function c(a){var c=new b();var d=[];if(!a||!a.length){c.done(d);return c;}var e=0;var f=a.length;function g(a){return function(){e+=1;d[a]=Array.prototype.slice.call(arguments);if(e===f)c.done(d);};}for(var h=0;h<f;h++)a[h].then(g(h));return c;}function d(a,c){var e=new b();if(a.length===0)e.done.apply(e,c);else a[0].apply(null,c).then(function(){a.splice(0,1);d(a,arguments).then(function(){e.done.apply(e,arguments);});});return e;}function e(a){var b="";if(typeof a==="string")b=a;else{var c=encodeURIComponent;for(var d in a)if(a.hasOwnProperty(d))b+='&'+c(d)+'='+c(a[d]);}return b;}function f(){var a;if(window.XMLHttpRequest)a=new XMLHttpRequest();else if(window.ActiveXObject)try{a=new ActiveXObject("Msxml2.XMLHTTP");}catch(b){a=new ActiveXObject("Microsoft.XMLHTTP");}return a;}function g(a,c,d,g){var h=new b();var j,k;d=d||{};g=g||{};try{j=f();}catch(l){h.done(i.ENOXHR,"");return h;}k=e(d);if(a==='GET'&&k){c+='?'+k;k=null;}j.open(a,c);j.setRequestHeader('Content-type','application/x-www-form-urlencoded');for(var m in g)if(g.hasOwnProperty(m))j.setRequestHeader(m,g[m]);function n(){j.abort();h.done(i.ETIMEOUT,"",j);}var o=i.ajaxTimeout;if(o)var p=setTimeout(n,o);j.onreadystatechange=function(){if(o)clearTimeout(p);if(j.readyState===4){var a=(!j.status||(j.status<200||j.status>=300)&&j.status!==304);h.done(a,j.responseText,j);}};j.send(k);return h;}function h(a){return function(b,c,d){return g(a,b,c,d);};}var i={Promise:b,join:c,chain:d,ajax:g,get:h('GET'),post:h('POST'),put:h('PUT'),del:h('DELETE'),ENOXHR:1,ETIMEOUT:2,ajaxTimeout:0};if(typeof define==='function'&&define.amd)define(function(){return i;});else a.promise=i;})(this);

function loadDisgo() {
  $.each($('[data-disgo-url]'), function(i, el) {
    appendForm(el)
    initializeComments(el)
  });

  function initializeComments(el) {
    var url = $(el).attr('data-disgo-url')
    promise.get('http://localhost:3000/comments', {url: url}).then(function(error, text, xhr) {
        if (error) {
            alert('Error ' + xhr.status);
            return;
        }
        var comments = JSON.parse(text)
        $.each(comments, function(i, comment) {
          appendComment(el, comment)
        })
    });
  }

  function appendForm(el) {
    var url = $(el).attr('data-disgo-url')
    var form = document.createElement('form')
    var body = document.createElement('textarea')
    var submit = document.createElement('button')
    var email = document.createElement('input')
    var urlField = document.createElement('input')
    urlField.setAttribute('type', 'hidden')
    urlField.setAttribute('name', 'url')
    urlField.setAttribute('value', url)
    body.setAttribute('name', 'body')
    email.setAttribute('type', 'text')
    email.setAttribute('name', 'email')
    submit.textContent = 'Comment'
    submit.addEventListener('click', function(e) {
      e.preventDefault()
      submitComment(el)
    })
    form.appendChild(email)
    form.appendChild(body)
    form.appendChild(submit)
    form.appendChild(urlField)
    el.appendChild(form)
  }

  function appendComment(el, comment) {
    var avatar = document.createElement('img')
    avatar.setAttribute('src', comment.avatar)
    var body = document.createElement('div')
    body.textContent = comment.body
    el.appendChild(body)
    el.appendChild(avatar)
  }

  function submitComment(el) {
    var form = $("form", el)
    promise.post('http://localhost:3000/comments', form.serialize()).then(function(error, text, xhr) {
        if (error) {
            alert('Error ' + xhr.status);
            return;
        }
        var comment = JSON.parse(text)
        appendComment(el, comment)
    })
  }
}

window[addEventListener ? 'addEventListener' : 'attachEvent'](addEventListener ? 'load' : 'onload', loadDisgo)
