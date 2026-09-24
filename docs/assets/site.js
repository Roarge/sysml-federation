// Click-to-define terms.
//
// An article marks a term as <span class="term" data-term="id">text</span>,
// which github.com shows as plain text. On the site the layout embeds the
// definitions from _data/glossary.yml as JSON, and this script turns each
// marked term into a button that opens its definition beside it. Without the
// script the words stay ordinary text, and nothing looks clickable that is not.
(function () {
  "use strict";

  var source = document.getElementById("glossary-data");
  if (!source) return;
  var entries;
  try {
    entries = JSON.parse(source.textContent);
  } catch (err) {
    return;
  }
  var byId = {};
  entries.forEach(function (entry) {
    byId[entry.id] = entry;
  });

  var open = null; // { button: HTMLElement, pop: HTMLElement }
  var serial = 0;

  function linkLabel(href) {
    try {
      var url = new URL(href);
      return url.hostname.replace(/^www\./, "");
    } catch (err) {
      return href;
    }
  }

  function build(entry, id) {
    var pop = document.createElement("div");
    pop.className = "term-pop";
    pop.id = id;
    pop.setAttribute("role", "dialog");
    pop.setAttribute("aria-label", entry.term);
    pop.tabIndex = -1;

    var head = document.createElement("div");
    head.className = "term-pop-head";
    var title = document.createElement("p");
    title.className = "term-pop-title";
    title.textContent = entry.term;
    var shut = document.createElement("button");
    shut.type = "button";
    shut.className = "term-pop-close";
    shut.setAttribute("aria-label", "Close definition");
    shut.innerHTML =
      '<svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" aria-hidden="true"><path d="M2 2 L12 12 M12 2 L2 12"/></svg>';
    shut.addEventListener("click", function () {
      close(true);
    });
    head.appendChild(title);
    head.appendChild(shut);
    pop.appendChild(head);

    var text = document.createElement("p");
    text.className = "term-pop-text";
    text.textContent = entry.definition;
    pop.appendChild(text);

    if (entry.link) {
      var link = document.createElement("a");
      link.className = "term-pop-link";
      link.href = entry.link;
      link.target = "_blank";
      link.rel = "noopener";
      link.textContent = linkLabel(entry.link);
      link.insertAdjacentHTML(
        "beforeend",
        '<svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 12 L12 4"/><path d="M6 4 H12 V10"/></svg>'
      );
      pop.appendChild(link);
    }
    return pop;
  }

  // Below the term, kept inside the viewport, with the caret under the
  // term's middle. The popover sits in the document, so it scrolls with it.
  function place(button, pop) {
    var rect = button.getBoundingClientRect();
    var width = pop.offsetWidth;
    var gutter = 16;
    var viewport = document.documentElement.clientWidth;
    var centre = rect.left + rect.width / 2;
    var left = Math.min(Math.max(centre - 40, gutter), viewport - width - gutter);
    pop.style.left = window.scrollX + left + "px";
    pop.style.top = window.scrollY + rect.bottom + 10 + "px";
    var caret = Math.min(Math.max(centre - left, 16), width - 16);
    pop.style.setProperty("--caret-x", caret + "px");
  }

  function close(returnFocus) {
    if (!open) return;
    var button = open.button;
    open.pop.remove();
    button.setAttribute("aria-expanded", "false");
    open = null;
    if (returnFocus) button.focus();
  }

  function show(button) {
    var entry = byId[button.getAttribute("data-term")];
    if (!entry) return;
    close(false);
    var id = "term-pop-" + ++serial;
    var pop = build(entry, id);
    document.body.appendChild(pop);
    button.setAttribute("aria-expanded", "true");
    button.setAttribute("aria-controls", id);
    place(button, pop);
    open = { button: button, pop: pop };
    pop.focus({ preventScroll: true });
    // The popover is appended at the end of the page, so tabbing out of it
    // would land past the footer. Tabbing off either end closes it instead
    // and hands focus back to the term, and reading order carries on there.
    pop.addEventListener("keydown", function (event) {
      if (event.key !== "Tab") return;
      var stops = pop.querySelectorAll("button, a[href]");
      var first = stops[0];
      var lastStop = stops[stops.length - 1];
      var leaving =
        (event.shiftKey && (document.activeElement === first || document.activeElement === pop)) ||
        (!event.shiftKey && document.activeElement === lastStop);
      if (leaving) {
        event.preventDefault();
        close(true);
      }
    });
  }

  var spans = document.querySelectorAll("span.term[data-term]");
  Array.prototype.forEach.call(spans, function (span) {
    var id = span.getAttribute("data-term");
    if (!byId[id]) return;
    var button = document.createElement("button");
    button.type = "button";
    button.className = "term";
    button.setAttribute("data-term", id);
    button.setAttribute("aria-expanded", "false");
    while (span.firstChild) button.appendChild(span.firstChild);
    span.parentNode.replaceChild(button, span);
    button.addEventListener("click", function (event) {
      event.stopPropagation();
      if (open && open.button === button) {
        close(true);
      } else {
        show(button);
      }
    });
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && open) close(true);
  });
  document.addEventListener("click", function (event) {
    if (open && !open.pop.contains(event.target)) close(false);
  });
  window.addEventListener("resize", function () {
    if (open) place(open.button, open.pop);
  });
})();
