package html

// interactivityJS returns the vanilla JS for blog interactivity.
const interactivityJS = `
(function() {
  'use strict';

  document.querySelectorAll('.fluffy-List li').forEach(function(li) {
    li.style.cursor = 'pointer';
    li.addEventListener('click', function() {
      document.querySelectorAll('.fluffy-List li').forEach(function(el) {
        el.classList.remove('selected');
      });
      li.classList.add('selected');
      var slug = li.getAttribute('data-slug');
      if (slug) {
        document.querySelectorAll('article[data-slug]').forEach(function(a) {
          a.style.display = a.getAttribute('data-slug') === slug ? '' : 'none';
        });
      }
    });
  });

  document.querySelectorAll('.fluffy-Chip').forEach(function(chip) {
    chip.addEventListener('click', function() {
      document.querySelectorAll('.fluffy-Chip').forEach(function(c) {
        c.classList.remove('active');
      });
      chip.classList.add('active');
      var tag = chip.textContent.trim();
      filterPosts(tag, getSearchQuery());
    });
  });

  var searchInput = document.querySelector('.fluffy-Search');
  if (searchInput) {
    searchInput.addEventListener('input', function() {
      var activeTag = getActiveTag();
      filterPosts(activeTag, searchInput.value);
    });
  }

  function getActiveTag() {
    var active = document.querySelector('.fluffy-Chip.active');
    return active ? active.textContent.trim() : 'ALL';
  }

  function getSearchQuery() {
    var input = document.querySelector('.fluffy-Search');
    return input ? input.value : '';
  }

  function filterPosts(tag, query) {
    var items = document.querySelectorAll('.fluffy-List li');
    var first = null;
    items.forEach(function(li) {
      var tags = (li.getAttribute('data-tags') || '').split(',');
      var title = li.textContent.toLowerCase();
      var matchTag = !tag || tag === 'ALL' || tags.indexOf(tag) >= 0;
      var matchQuery = !query || title.indexOf(query.toLowerCase()) >= 0;
      var visible = matchTag && matchQuery;
      li.style.display = visible ? '' : 'none';
      if (visible && !first) first = li;
    });
    if (first) first.click();
  }

  var firstItem = document.querySelector('.fluffy-List li');
  if (firstItem) firstItem.classList.add('selected');
})();
`
