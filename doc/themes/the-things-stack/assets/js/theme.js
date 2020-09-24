// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

document.addEventListener('DOMContentLoaded', function () {
  var $navbarBurgers = Array.prototype.slice.call(document.querySelectorAll('.navbar-burger'), 0);
  if ($navbarBurgers.length > 0) {
    $navbarBurgers.forEach(function (el) {
      el.addEventListener('click', function () {
        var target = el.dataset.target;
        var $target = document.getElementById(target);
        el.classList.toggle('is-active');
        $target.classList.toggle('is-active');
      });
    });
  }

  const tabs = [...document.querySelectorAll('.tabs li')];
  const tabContent = [...document.querySelectorAll('.tab-content section')];
  const activeClass = 'is-active';

  function init() {
    if(window.sessionStorage.getItem('tabActive')) {
      toggleTab(window.sessionStorage.getItem('tabActive'))
    }
    else if(tabs){
      let key = tabs[0].getAttribute('data-tab');
      toggleTab(key)
    }

    tabs.forEach((tab) => {
      tab.addEventListener('click', handleClick)
    })
  }

  function handleClick(event) {
    event.preventDefault();
    let tab = event.currentTarget;
    let key = tab.getAttribute('data-tab');
    toggleTab(key)
  }

  function toggleTab(key) {
    let activeTabs = getTabsByKey(key)

    if(!activeTabs) return

    tabs.forEach((tab) => {
      if (tab && tab.classList.contains(activeClass)) {
        tab.classList.remove(activeClass);
      }
    });

    activeTabs.forEach((tab) => {
      tab.classList.add(activeClass);
    })

    tabContent.forEach((item) => {
      if (item && item.classList.contains(activeClass)) {
        item.classList.remove(activeClass);
      }
      let data = item.getAttribute('data-content');
      if (data === key) {
        item.classList.add(activeClass);
      }
    });
    if(window.sessionStorage){
      window.sessionStorage.setItem("tabActive", key)
    }
  }

  function getTabsByKey(key) {
    return [...document.querySelectorAll(`[data-tab="${key}"]`)];
  }

  init();

});
