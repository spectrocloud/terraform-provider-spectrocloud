(function () {
  var input = document.getElementById("search-input");
  var results = document.getElementById("search-results");
  if (!input || !results || typeof SEARCH_INDEX === "undefined") return;

  var relRoot = document.body.getAttribute("data-rel-root") || "";

  function render(matches, query) {
    if (!query) {
      results.classList.remove("open");
      results.innerHTML = "";
      return;
    }
    if (matches.length === 0) {
      results.innerHTML = '<div class="no-results">No matches for "' + escapeHtml(query) + '"</div>';
      results.classList.add("open");
      return;
    }
    var html = matches
      .slice(0, 20)
      .map(function (item) {
        return (
          '<a href="' + relRoot + item.url + '">' +
          '<div class="result-title">' + escapeHtml(item.title) + "</div>" +
          '<div class="result-meta">' + escapeHtml(item.section) + "</div>" +
          "</a>"
        );
      })
      .join("");
    results.innerHTML = html;
    results.classList.add("open");
  }

  function escapeHtml(s) {
    return s.replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  function search(query) {
    var q = query.trim().toLowerCase();
    if (!q) return [];
    return SEARCH_INDEX.filter(function (item) {
      return (
        item.title.toLowerCase().indexOf(q) !== -1 ||
        item.snippet.toLowerCase().indexOf(q) !== -1 ||
        item.section.toLowerCase().indexOf(q) !== -1
      );
    });
  }

  input.addEventListener("input", function () {
    render(search(input.value), input.value);
  });

  input.addEventListener("focus", function () {
    if (input.value.trim()) render(search(input.value), input.value);
  });

  document.addEventListener("click", function (e) {
    if (!e.target.closest(".search-box")) {
      results.classList.remove("open");
    }
  });

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape") {
      results.classList.remove("open");
      input.blur();
    }
  });
})();
