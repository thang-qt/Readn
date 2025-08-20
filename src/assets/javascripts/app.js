'use strict';

var TITLE = document.title

function scrollto(target, scroll) {
  var padding = 10
  var targetRect = target.getBoundingClientRect()
  var scrollRect = scroll.getBoundingClientRect()

  // target
  var relativeOffset = targetRect.y - scrollRect.y
  var absoluteOffset = relativeOffset + scroll.scrollTop

  if (padding <= relativeOffset && relativeOffset + targetRect.height <= scrollRect.height - padding) return

  var newPos = scroll.scrollTop
  if (relativeOffset < padding) {
    newPos = absoluteOffset - padding
  } else {
    newPos = absoluteOffset - scrollRect.height + targetRect.height + padding
  }
  scroll.scrollTop = Math.round(newPos)
}

var debounce = function(callback, wait) {
  var timeout
  return function() {
    var ctx = this, args = arguments
    clearTimeout(timeout)
    timeout = setTimeout(function() {
      callback.apply(ctx, args)
    }, wait)
  }
}

Vue.directive('scroll', {
  inserted: function(el, binding) {
    el.addEventListener('scroll', debounce(function(event) {
      binding.value(event, el)
    }, 200))
  },
})

Vue.directive('focus', {
  inserted: function(el) {
    el.focus()
  }
})

Vue.component('drag', {
  props: ['width'],
  template: '<div class="drag"></div>',
  mounted: function() {
    var self = this
    var startX = undefined
    var initW = undefined
    var onMouseMove = function(e) {
      var offset = e.clientX - startX
      var newWidth = initW + offset
      self.$emit('resize', newWidth)
    }
    var onMouseUp = function(e) {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
    }
    this.$el.addEventListener('mousedown', function(e) {
      startX = e.clientX
      initW = self.width
      document.addEventListener('mousemove', onMouseMove)
      document.addEventListener('mouseup', onMouseUp)
    })
  },
})

Vue.component('dropdown', {
  props: ['class', 'toggle-class', 'ref', 'drop', 'title'],
  data: function() {
    return {open: false}
  },
  template: `
    <div class="dropdown" :class="$attrs.class">
      <button ref="btn" @click="toggle" :class="btnToggleClass" :title="$props.title"><slot name="button"></slot></button>
      <div ref="menu" class="dropdown-menu" :class="{show: open}"><slot v-if="open"></slot></div>
    </div>
  `,
  computed: {
    btnToggleClass: function() {
      var c = this.$props.toggleClass || ''
      c += ' dropdown-toggle dropdown-toggle-no-caret'
      c += this.open ? ' show' : ''
      return c.trim()
    }
  },
  methods: {
    toggle: function(e) {
      this.open ? this.hide() : this.show()
    },
    show: function(e) {
      this.open = true
      this.$refs.menu.style.top = this.$refs.btn.offsetHeight + 'px'
      var drop = this.$props.drop

      if (drop === 'right') {
        this.$refs.menu.style.left = 'auto'
        this.$refs.menu.style.right = '0'
      } else if (drop === 'center') {
        this.$nextTick(function() {
          var btnWidth = this.$refs.btn.getBoundingClientRect().width
          var menuWidth = this.$refs.menu.getBoundingClientRect().width
          this.$refs.menu.style.left = '-' + ((menuWidth - btnWidth) / 2) + 'px'
        }.bind(this))
      }

      document.addEventListener('click', this.clickHandler)
    },
    hide: function() {
      this.open = false
      document.removeEventListener('click', this.clickHandler)
    },
    clickHandler: function(e) {
      var dropdown = e.target.closest('.dropdown')
      if (dropdown == null || dropdown != this.$el) return this.hide()
      if (e.target.closest('.dropdown-item') != null) return this.hide()
    }
  },
})

Vue.component('modal', {
  props: ['open'],
  template: `
    <div class="modal custom-modal" tabindex="-1" v-if="$props.open">
      <div class="modal-dialog">
        <div class="modal-content" ref="content">
          <div class="modal-body">
            <slot v-if="$props.open"></slot>
          </div>
        </div>
      </div>
    </div>
  `,
  data: function() {
    return {opening: false}
  },
  watch: {
    'open': function(newVal) {
      if (newVal) {
        this.opening = true
        document.addEventListener('click', this.handleClick)
      } else {
        document.removeEventListener('click', this.handleClick)
      }
    },
  },
  methods: {
    handleClick: function(e) {
      if (this.opening) {
        this.opening = false
        return
      }
      if (e.target.closest('.modal-content') == null) this.$emit('hide')
    },
  },
})

function dateRepr(d) {
  var sec = (new Date().getTime() - d.getTime()) / 1000
  var neg = sec < 0
  var out = ''

  sec = Math.abs(sec)
  if (sec < 2700)  // less than 45 minutes
    out = Math.round(sec / 60) + 'm'
  else if (sec < 86400)  // less than 24 hours
    out = Math.round(sec / 3600) + 'h'
  else if (sec < 604800)  // less than a week
    out = Math.round(sec / 86400) + 'd'
  else
    out = d.toLocaleDateString(undefined, {year: "numeric", month: "long", day: "numeric"})

  if (neg) return '-' + out
  return out
}

Vue.component('relative-time', {
  props: ['val'],
  data: function() {
    var d = new Date(this.val)
    return {
      'date': d,
      'formatted': dateRepr(d),
      'interval': null,
    }
  },
  template: '<time :datetime="val">{{ formatted }}</time>',
  mounted: function() {
    this.interval = setInterval(function() {
      this.formatted = dateRepr(this.date)
    }.bind(this), 600000)  // every 10 minutes
  },
  destroyed: function() {
    clearInterval(this.interval)
  },
})

var vm = new Vue({
  created: function() {
    this.refreshStats()
      .then(this.refreshFeeds.bind(this))
      .then(this.refreshItems.bind(this, false))

    api.feeds.list_errors().then(function(errors) {
      vm.feed_errors = errors
    })
  },
  mounted: function() {
    this.initTextSelection()
  },
  data: function() {
    var s = app.settings
    return {
      'filterSelected': s.filter,
      'folders': [],
      'feeds': [],
      'feedSelected': s.feed,
      'feedListWidth': s.feed_list_width || 300,
      'feedNewChoice': [],
      'feedNewChoiceSelected': '',
      'items': [],
      'itemsHasMore': true,
      'itemSelected': null,
      'itemSelectedDetails': null,
      'itemSelectedReadability': '',
      'itemSelectedHNDiscussion': '',
      'itemSelectedSummary': '',
      'summaryError': '',
      'feedSummary': '',
      'feedSummaryError': '',
      'feedSummaryTitle': '',
      'itemSearch': '',
      'itemSortNewestFirst': s.sort_newest_first,
      'itemListWidth': s.item_list_width || 300,

      'filteredFeedStats': {},
      'filteredFolderStats': {},
      'filteredTotalStats': null,

      'settings': '',
      'loading': {
        'feeds': 0,
        'newfeed': false,
        'items': false,
        'readability': false,
        'hnDiscussion': false,
        'summary': false,
        'feedSummary': false,
        'chat': false,
      },
      'fonts': ['', 'serif', 'monospace'],
      'feedStats': {},
      'theme': {
        'name': s.theme_name,
        'font': s.theme_font,
        'size': s.theme_size,
      },
      'refreshRate': s.refresh_rate,
      'aiKey': s.ai_api_key || '',
      'aiURL': s.ai_api_url || 'https://api.aimlapi.com/v1/chat/completions',
      'aiModel': s.ai_model || 'gpt-4o-mini',
      'aiPrompt': s.ai_prompt || 'Please provide a concise summary (TL;DR) of the following article. Keep summaries between 2-4 sentences, highlighting the key points and important details:',
      'aiPersonality': s.ai_personality || 'You are a helpful, knowledgeable assistant that provides clear and concise responses.',
      'aiExplainPrompt': s.ai_explain_prompt || 'Please explain this text in a clear and easy-to-understand way:',
      'aiSummarizePrompt': s.ai_summarize_prompt || 'Please provide a concise summary of this text:',
      'authenticated': app.authenticated,
      'feed_errors': {},
      'sidebarCollapsed': s.sidebar_collapsed,
      'chatPanelVisible': false,
      'chatMessages': [],
      'chatInput': '',
      'chatContext': '',
    }
  },
  computed: {
    foldersWithFeeds: function() {
      var feedsByFolders = this.feeds.reduce(function(folders, feed) {
        if (!folders[feed.folder_id])
          folders[feed.folder_id] = [feed]
        else
          folders[feed.folder_id].push(feed)
        return folders
      }, {})
      var folders = this.folders.slice().map(function(folder) {
        folder.feeds = feedsByFolders[folder.id]
        return folder
      })
      folders.push({id: null, feeds: feedsByFolders[null]})
      return folders
    },
    feedsById: function() {
      return this.feeds.reduce(function(acc, f) { acc[f.id] = f; return acc }, {})
    },
    foldersById: function() {
      return this.folders.reduce(function(acc, f) { acc[f.id] = f; return acc }, {})
    },
    current: function() {
      var parts = (this.feedSelected || '').split(':', 2)
      var type = parts[0]
      var guid = parts[1]

      var folder = {}, feed = {}

      if (type == 'feed')
        feed = this.feedsById[guid] || {}
      if (type == 'folder')
        folder = this.foldersById[guid] || {}

      return {type: type, feed: feed, folder: folder}
    },
    itemSelectedContent: function() {
      if (!this.itemSelected) return ''

      if (this.itemSelectedHNDiscussion)
        return this.itemSelectedHNDiscussion

      if (this.itemSelectedReadability)
        return this.itemSelectedReadability

      return this.itemSelectedDetails.content || ''
    },
    contentImages: function() {
      if (!this.itemSelectedDetails) return []
      return (this.itemSelectedDetails.media_links || []).filter(l => l.type === 'image')
    },
    contentAudios: function() {
      if (!this.itemSelectedDetails) return []
      return (this.itemSelectedDetails.media_links || []).filter(l => l.type === 'audio')
    },
    contentVideos: function() {
      if (!this.itemSelectedDetails) return []
      return (this.itemSelectedDetails.media_links || []).filter(l => l.type === 'video')
    }
  },
  watch: {
    'theme': {
      deep: true,
      handler: function(theme) {
        document.body.classList.value = 'theme-' + theme.name
        api.settings.update({
          theme_name: theme.name,
          theme_font: theme.font,
          theme_size: theme.size,
        })
      },
    },
    'feedStats': {
      deep: true,
      handler: debounce(function() {
        var title = TITLE
        var unreadCount = Object.values(this.feedStats).reduce(function(acc, stat) {
          return acc + stat.unread
        }, 0)
        if (unreadCount) {
          title += ' ('+unreadCount+')'
        }
        document.title = title
        this.computeStats()
      }, 500),
    },
    'filterSelected': function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({filter: newVal}).then(this.refreshItems.bind(this, false))
      this.itemSelected = null
      // Clear feed summary when changing filter
      this.feedSummary = ''
      this.feedSummaryError = ''
      this.computeStats()
    },
    'feedSelected': function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({feed: newVal}).then(this.refreshItems.bind(this, false))
      this.itemSelected = null
      // Clear feed summary when changing feeds
      this.feedSummary = ''
      this.feedSummaryError = ''
      if (this.$refs.itemlist) this.$refs.itemlist.scrollTop = 0
    },
    'itemSelected': function(newVal, oldVal) {
      this.itemSelectedReadability = ''
      this.itemSelectedSummary = ''
      this.summaryError = ''
      // Clear feed summary when selecting an article
      this.feedSummary = ''
      this.feedSummaryError = ''
      if (newVal === null) {
        this.itemSelectedDetails = null
        return
      }
      if (this.$refs.content) this.$refs.content.scrollTop = 0

      api.items.get(newVal).then(function(item) {
        this.itemSelectedDetails = item
        if (this.itemSelectedDetails.status == 'unread') {
          api.items.update(this.itemSelectedDetails.id, {status: 'read'}).then(function() {
            this.feedStats[this.itemSelectedDetails.feed_id].unread -= 1
            var itemInList = this.items.find(function(i) { return i.id == item.id })
            if (itemInList) itemInList.status = 'read'
            this.itemSelectedDetails.status = 'read'
          }.bind(this))
        }
      }.bind(this))
    },
    'itemSearch': debounce(function(newVal) {
      this.refreshItems()
    }, 500),
    'itemSortNewestFirst': function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({sort_newest_first: newVal}).then(vm.refreshItems.bind(this, false))
    },
    'feedListWidth': debounce(function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({feed_list_width: newVal})
    }, 1000),
    'itemListWidth': debounce(function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({item_list_width: newVal})
    }, 1000),
    'refreshRate': function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({refresh_rate: newVal})
    },
    'sidebarCollapsed': function(newVal, oldVal) {
      if (oldVal === undefined) return  // do nothing, initial setup
      api.settings.update({sidebar_collapsed: newVal})
    },
    'itemSelected': function(newVal, oldVal) {
      this.itemSelectedReadability = ''
      this.itemSelectedHNDiscussion = ''
      this.itemSelectedSummary = ''
      this.summaryError = ''
      // Clear chat when switching articles
      if (oldVal !== undefined && newVal !== oldVal) {
        this.chatMessages = []
      }
      if (newVal === null) {
        this.itemSelectedDetails = null
        return
      }
      if (this.$refs.content) this.$refs.content.scrollTop = 0

      api.items.get(newVal).then(function(item) {
        vm.itemSelectedDetails = item
        if (vm.itemSelectedDetails.status == 'unread') {
          api.items.update(vm.itemSelectedDetails.id, {status: 'read'}).then(function() {
            vm.feedStats[vm.itemSelectedDetails.feed_id].unread -= 1
            var itemInList = vm.items.find(function(i) { return i.id == item.id })
            if (itemInList) itemInList.status = 'read'
            vm.itemSelectedDetails.status = 'read'
          })
        }
      })
    },
    'apiKey': function(newVal, oldVal) {
      if (oldVal === undefined) return
      api.settings.update({summary_api_key: newVal})
    },
  },
  methods: {
    refreshStats: function(loopMode) {
      return api.status().then(function(data) {
        if (loopMode && !vm.itemSelected) vm.refreshItems()

        vm.loading.feeds = data.running
        if (data.running) {
          setTimeout(vm.refreshStats.bind(vm, true), 500)
        }
        vm.feedStats = data.stats.reduce(function(acc, stat) {
          acc[stat.feed_id] = stat
          return acc
        }, {})

        api.feeds.list_errors().then(function(errors) {
          vm.feed_errors = errors
        })
      })
    },
    getItemsQuery: function() {
      var query = {}
      if (this.feedSelected) {
        var parts = this.feedSelected.split(':', 2)
        var type = parts[0]
        var guid = parts[1]
        if (type == 'feed') {
          query.feed_id = guid
        } else if (type == 'folder') {
          query.folder_id = guid
        }
      }
      if (this.filterSelected) {
        query.status = this.filterSelected
      }
      if (this.itemSearch) {
        query.search = this.itemSearch
      }
      if (!this.itemSortNewestFirst) {
        query.oldest_first = true
      }
      return query
    },
    refreshFeeds: function() {
      return Promise
        .all([api.folders.list(), api.feeds.list()])
        .then(function(values) {
          vm.folders = values[0]
          vm.feeds = values[1]
        })
    },
    refreshItems: function(loadMore = false) {
      if (this.feedSelected === null) {
        vm.items = []
        vm.itemsHasMore = false
        return
      }

      var query = this.getItemsQuery()
      if (loadMore) {
        query.after = vm.items[vm.items.length-1].id
      }

      this.loading.items = true
      return api.items.list(query).then(function(data) {
        if (loadMore) {
          vm.items = vm.items.concat(data.list)
        } else {
          vm.items = data.list
        }
        vm.itemsHasMore = data.has_more
        vm.loading.items = false

        // load more if there's some space left at the bottom of the item list.
        vm.$nextTick(function() {
          if (vm.itemsHasMore && !vm.loading.items && vm.itemListCloseToBottom()) {
            vm.refreshItems(true)
          }
        })
      })
    },
    itemListCloseToBottom: function() {
      // approx. vertical space at the bottom of the list (loading el & paddings) when 1rem = 16px
      var bottomSpace = 70
      var scale = (parseFloat(getComputedStyle(document.documentElement).fontSize) || 16) / 16

      var el = this.$refs.itemlist

      if (el.scrollHeight === 0) return false  // element is invisible (responsive design)

      var closeToBottom = (el.scrollHeight - el.scrollTop - el.offsetHeight) < bottomSpace * scale
      return closeToBottom
    },
    loadMoreItems: function(event, el) {
      if (!this.itemsHasMore) return
      if (this.loading.items) return
      if (this.itemListCloseToBottom()) return this.refreshItems(true)
      if (this.itemSelected && this.itemSelected === this.items[this.items.length - 1].id) return this.refreshItems(true)
    },
    markItemsRead: function() {
      var query = this.getItemsQuery()
      api.items.mark_read(query).then(function() {
        vm.items = []
        vm.itemsPage = {'cur': 1, 'num': 1}
        vm.itemSelected = null
        vm.itemsHasMore = false
        vm.refreshStats()
      })
    },
    toggleFolderExpanded: function(folder) {
      folder.is_expanded = !folder.is_expanded
      api.folders.update(folder.id, {is_expanded: folder.is_expanded})
    },
    formatDate: function(datestr) {
      var options = {
        year: "numeric", month: "long", day: "numeric",
        hour: '2-digit', minute: '2-digit',
      }
      return new Date(datestr).toLocaleDateString(undefined, options)
    },
    moveFeed: function(feed, folder) {
      var folder_id = folder ? folder.id : null
      api.feeds.update(feed.id, {folder_id: folder_id}).then(function() {
        feed.folder_id = folder_id
        vm.refreshStats()
      })
    },
    moveFeedToNewFolder: function(feed) {
      var title = prompt('Enter folder name:')
      if (!title) return
      api.folders.create({'title': title}).then(function(folder) {
        api.feeds.update(feed.id, {folder_id: folder.id}).then(function() {
          vm.refreshFeeds().then(function() {
            vm.refreshStats()
          })
        })
      })
    },
    createNewFeedFolder: function() {
      var title = prompt('Enter folder name:')
      if (!title) return
      api.folders.create({'title': title}).then(function(result) {
        vm.refreshFeeds().then(function() {
          vm.$nextTick(function() {
            if (vm.$refs.newFeedFolder) {
              vm.$refs.newFeedFolder.value = result.id
            }
          })
        })
      })
    },
    renameFolder: function(folder) {
      var newTitle = prompt('Enter new title', folder.title)
      if (newTitle) {
        api.folders.update(folder.id, {title: newTitle}).then(function() {
          folder.title = newTitle
          this.folders.sort(function(a, b) {
            return a.title.localeCompare(b.title)
          })
        }.bind(this))
      }
    },
    deleteFolder: function(folder) {
      if (confirm('Are you sure you want to delete ' + folder.title + '?')) {
        api.folders.delete(folder.id).then(function() {
          vm.feedSelected = null
          vm.refreshStats()
          vm.refreshFeeds()
        })
      }
    },
    updateFeedLink: function(feed) {
      var newLink = prompt('Enter feed link', feed.feed_link)
      if (newLink) {
        api.feeds.update(feed.id, {feed_link: newLink}).then(function() {
          feed.feed_link = newLink
        })
      }
    },
    renameFeed: function(feed) {
      var newTitle = prompt('Enter new title', feed.title)
      if (newTitle) {
        api.feeds.update(feed.id, {title: newTitle}).then(function() {
          feed.title = newTitle
        })
      }
    },
    deleteFeed: function(feed) {
      if (confirm('Are you sure you want to delete ' + feed.title + '?')) {
        api.feeds.delete(feed.id).then(function() {
          vm.feedSelected = null
          vm.refreshStats()
          vm.refreshFeeds()
        })
      }
    },
    createFeed: function(event) {
      var form = event.target
      var data = {
        url: form.querySelector('input[name=url]').value,
        folder_id: parseInt(form.querySelector('select[name=folder_id]').value) || null,
      }
      if (this.feedNewChoiceSelected) {
        data.url = this.feedNewChoiceSelected
      }
      this.loading.newfeed = true
      api.feeds.create(data).then(function(result) {
        if (result.status === 'success') {
          vm.refreshFeeds()
          vm.refreshStats()
          vm.settings = ''
          vm.feedSelected = 'feed:' + result.feed.id
        } else if (result.status === 'multiple') {
          vm.feedNewChoice = result.choice
          vm.feedNewChoiceSelected = result.choice[0].url
        } else {
          alert('No feeds found at the given url.')
        }
        vm.loading.newfeed = false
      })
    },
    toggleItemStatus: function(item, targetstatus, fallbackstatus) {
      var oldstatus = item.status
      var newstatus = item.status !== targetstatus ? targetstatus : fallbackstatus

      var updateStats = function(status, incr) {
        if ((status == 'unread') || (status == 'starred')) {
          this.feedStats[item.feed_id][status] += incr
        }
      }.bind(this)

      api.items.update(item.id, {status: newstatus}).then(function() {
        updateStats(oldstatus, -1)
        updateStats(newstatus, +1)

        var itemInList = this.items.find(function(i) { return i.id == item.id })
        if (itemInList) itemInList.status = newstatus
        item.status = newstatus
      }.bind(this))
    },
    toggleItemStarred: function(item) {
      this.toggleItemStatus(item, 'starred', 'read')
    },
    toggleItemRead: function(item) {
      this.toggleItemStatus(item, 'unread', 'read')
    },
    importOPML: function(event) {
      var input = event.target
      var form = document.querySelector('#opml-import-form')
      this.$refs.menuDropdown.hide()
      api.upload_opml(form).then(function() {
        input.value = ''
        vm.refreshFeeds()
        vm.refreshStats()
      })
    },
    logout: function() {
      api.logout().then(function() {
        document.location.reload()
      })
    },
    toggleReadability: function() {
      if (this.itemSelectedReadability) {
        this.itemSelectedReadability = null
        return
      }
      var item = this.itemSelectedDetails
      if (!item) return
      if (item.link) {
        this.loading.readability = true
        api.crawl(item.link).then(function(data) {
          vm.itemSelectedReadability = data && data.content
          vm.loading.readability = false
        })
      }
    },
    isHackerNewsItem: function(item) {
      if (!item) return false
      var isHNFeed = item.link && (item.link.includes('news.ycombinator.com') || item.link.includes('ycombinator.com'))
      var hasHNDiscussion = item.content && item.content.includes('news.ycombinator.com/item?id=')
      return isHNFeed || hasHNDiscussion
    },
    toggleHNDiscussion: function() {
      if (this.itemSelectedHNDiscussion) {
        this.itemSelectedHNDiscussion = ''
        return
      }
      var item = this.itemSelectedDetails
      if (!item) return
      
      this.loading.hnDiscussion = true
      var vm = this
      
      api.hackernews({
        content: item.content || '',
        url: item.link || ''
      }).then(function(data) {
        vm.itemSelectedHNDiscussion = data && data.html
        vm.loading.hnDiscussion = false
      }).catch(function(error) {
        console.error('Error fetching HN discussion:', error)
        vm.loading.hnDiscussion = false
      })
    },
    toggleSummary: function() {
      if (this.itemSelectedSummary) {
        this.itemSelectedSummary = ''
        this.summaryError = ''
        return
      }
      var item = this.itemSelectedDetails
      if (!item) return
      
      var content = this.itemSelectedReadability || item.content || ''
      if (!content) {
        this.summaryError = 'No content available to summarize'
        return
      }
      
      this.loading.summary = true
      this.summaryError = ''
      
      api.summarize(content, item.title).then(function(data) {
        vm.loading.summary = false
        if (data.error) {
          vm.summaryError = data.error
        } else {
          vm.itemSelectedSummary = data.summary
        }
      }).catch(function(error) {
        vm.loading.summary = false
        vm.summaryError = 'Failed to generate summary: ' + error.message
      })
    },
    updateAIKey: function(value) {
      this.aiKey = value
      api.settings.update({ai_api_key: value})
    },
    updateAIURL: function(value) {
      this.aiURL = value
      api.settings.update({ai_api_url: value})
    },
    updateAIModel: function(value) {
      this.aiModel = value
      api.settings.update({ai_model: value})
    },
    updateAIPrompt: function(value) {
      this.aiPrompt = value
      api.settings.update({ai_prompt: value})
    },
    updateAIPersonality: function(value) {
      this.aiPersonality = value
      api.settings.update({ai_personality: value})
    },
    updateAIExplainPrompt: function(value) {
      this.aiExplainPrompt = value
      api.settings.update({ai_explain_prompt: value})
    },
    updateAISummarizePrompt: function(value) {
      this.aiSummarizePrompt = value
      api.settings.update({ai_summarize_prompt: value})
    },
    summarizeFeed: function() {
      var query = this.getItemsQuery()
      this.loading.feedSummary = true
      this.feedSummaryError = ''
      // Clear previous feed summary content
      this.feedSummary = ''
      
      // Convert string IDs to numbers for API
      var folder_id = query.folder_id ? parseInt(query.folder_id) : null
      var feed_id = query.feed_id ? parseInt(query.feed_id) : null
      
      api.summarize_feed(folder_id, feed_id, query.status, query.search).then(function(data) {
        vm.loading.feedSummary = false
        if (data.error) {
          vm.feedSummaryError = data.error
          vm.feedSummary = ''
        } else {
          // Clear current selection and show feed summary
          vm.itemSelected = null
          vm.itemSelectedDetails = null
          vm.feedSummaryTitle = data.feed_title + ' - News Briefing (' + data.article_count + ' articles)'
          vm.feedSummary = data.summary
          vm.feedSummaryError = ''
        }
      }).catch(function(error) {
        vm.loading.feedSummary = false
        vm.feedSummaryError = 'Failed to generate feed summary: ' + error.message
        vm.feedSummary = ''
      })
    },
    toggleChatPanel: function() {
      this.chatPanelVisible = !this.chatPanelVisible
      if (this.chatPanelVisible && this.chatMessages.length === 0) {
        // Clear chat when opening panel
        this.clearChat()
      }
      if (this.chatPanelVisible) {
        this.$nextTick(function() {
          var chatContainer = vm.$refs.chatContainer
          if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight
          }
        })
      }
    },
    sendChatMessage: function() {
      if (!this.chatInput.trim() || this.loading.chat || !this.itemSelectedDetails) return
      
      var messageContent = this.chatInput.trim()
      
      // If there's context, include it in the message
      if (this.chatContext) {
        messageContent += '\n\nSelected text: "' + this.chatContext + '"'
      }
      
      var userMessage = {
        role: 'user',
        content: messageContent
      }
      
      this.chatMessages.push({
        role: 'user',
        content: this.chatInput.trim(),
        hasContext: !!this.chatContext,
        context: this.chatContext // Store the actual context text
      })
      
      var messages = this.chatMessages.slice() // Copy for API call
      // Replace the last message with the full context version for API
      if (this.chatContext) {
        messages[messages.length - 1] = userMessage
      }
      
      this.chatInput = ''
      this.chatContext = '' // Clear context after sending
      this.loading.chat = true
      
      var item = this.itemSelectedDetails
      var content = this.itemSelectedReadability || item.content || ''
      
      this.$nextTick(function() {
        var chatContainer = vm.$refs.chatContainer
        if (chatContainer) {
          chatContainer.scrollTop = chatContainer.scrollHeight
        }
      })
      
      api.chat(messages, item.title, content).then(function(data) {
        vm.loading.chat = false
        if (data.error) {
          vm.chatMessages.push({
            role: 'assistant',
            content: 'Sorry, I encountered an error: ' + data.error
          })
        } else {
          vm.chatMessages.push({
            role: 'assistant',
            content: data.response
          })
        }
        vm.$nextTick(function() {
          var chatContainer = vm.$refs.chatContainer
          if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight
          }
        })
      }).catch(function(error) {
        vm.loading.chat = false
        vm.chatMessages.push({
          role: 'assistant',
          content: 'Sorry, I encountered an error: ' + error.message
        })
        vm.$nextTick(function() {
          var chatContainer = vm.$refs.chatContainer
          if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight
          }
        })
      })
    },
    clearChat: function() {
      this.chatMessages = []
    },
    formatChatMessage: function(message) {
      // Use marked.js to render markdown
      if (typeof marked !== 'undefined') {
        // Convert literal \n strings to actual line breaks for markdown processing
        var processedMessage = message.replace(/\\n/g, '\n')
        // For markdown, we need double line breaks for paragraphs, single for line breaks
        // Replace single \n with double space + \n (markdown line break)
        processedMessage = processedMessage.replace(/\n(?!\n)/g, '  \n')
        return marked.parse(processedMessage)
      }
      // Fallback to basic formatting
      return message.replace(/\\n/g, '<br>').replace(/\n/g, '<br>')
    },
    showSettings: function(settings) {
      this.settings = settings

      if (settings === 'create') {
        vm.feedNewChoice = []
        vm.feedNewChoiceSelected = ''
      }
    },
    resizeFeedList: function(width) {
      this.feedListWidth = Math.min(Math.max(200, width), 700)
    },
    resizeItemList: function(width) {
      this.itemListWidth = Math.min(Math.max(200, width), 700)
    },
    resetFeedChoice: function() {
      this.feedNewChoice = []
      this.feedNewChoiceSelected = ''
    },
    incrFont: function(x) {
      this.theme.size = +(this.theme.size + (0.1 * x)).toFixed(1)
    },
    fetchAllFeeds: function() {
      if (this.loading.feeds) return
      api.feeds.refresh().then(function() {
        vm.refreshStats()
      })
    },
    computeStats: function() {
      var filter = this.filterSelected
      if (!filter) {
        this.filteredFeedStats = {}
        this.filteredFolderStats = {}
        this.filteredTotalStats = null
        return
      }

      var statsFeeds = {}, statsFolders = {}, statsTotal = 0

      for (var i = 0; i < this.feeds.length; i++) {
        var feed = this.feeds[i]
        if (!this.feedStats[feed.id]) continue

        var n = vm.feedStats[feed.id][filter] || 0

        if (!statsFolders[feed.folder_id]) statsFolders[feed.folder_id] = 0

        statsFeeds[feed.id] = n
        statsFolders[feed.folder_id] += n
        statsTotal += n
      }

      this.filteredFeedStats = statsFeeds
      this.filteredFolderStats = statsFolders
      this.filteredTotalStats = statsTotal
    },
    // navigation helper, navigate relative to selected item
    navigateToItem: function(relativePosition) {
      let vm = this
      if (vm.itemSelected == null) {
        // if no item is selected, select first
        if (vm.items.length !== 0) vm.itemSelected = vm.items[0].id
        return
      }

      var itemPosition = vm.items.findIndex(function(x) { return x.id === vm.itemSelected })
      if (itemPosition === -1) {
        if (vm.items.length !== 0) vm.itemSelected = vm.items[0].id
        return
      }

      var newPosition = itemPosition + relativePosition
      if (newPosition < 0 || newPosition >= vm.items.length) return

      vm.itemSelected = vm.items[newPosition].id

      vm.$nextTick(function() {
        var scroll = document.querySelector('#item-list-scroll')

        var handle = scroll.querySelector('input[type=radio]:checked')
        var target = handle && handle.parentElement

        if (target && scroll) scrollto(target, scroll)

        vm.loadMoreItems()
      })
    },
    // navigation helper, navigate relative to selected feed
    navigateToFeed: function(relativePosition) {
      let vm = this
      var navigationList = Array.from(document.querySelectorAll('#col-feed-list input[name=feed]'))
        .filter(function(r) { return r.offsetParent !== null && r.value !== 'folder:null' })
        .map(function(r) { return r.value })

      var currentFeedPosition = navigationList.indexOf(vm.feedSelected)

      if (currentFeedPosition == -1) {
        vm.feedSelected = ''
        return
      }

      var newPosition = currentFeedPosition+relativePosition
      if (newPosition < 0 || newPosition >= navigationList.length) return

      vm.feedSelected = navigationList[newPosition]

      vm.$nextTick(function() {
        var scroll = document.querySelector('#feed-list-scroll')

        var handle = scroll.querySelector('input[type=radio]:checked')
        var target = handle && handle.parentElement

        if (target && scroll) scrollto(target, scroll)
      })
    },
    toggleSidebarCollapsed: function() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    },
    
    // Text selection for AI chat functionality
    initTextSelection: function() {
      var self = this
      var selectedText = ''
      var tooltipJustShown = false
      
      // Handle text selection
      document.addEventListener('mouseup', function(e) {
        var tooltip = document.getElementById('ai-tooltip')
        if (!tooltip) return
        
        var selection = window.getSelection()
        if (!selection.rangeCount || selection.isCollapsed) {
          tooltip.classList.remove('show')
          return
        }
        
        var range = selection.getRangeAt(0)
        var contentArea = document.querySelector('.content')
        
        // Check if selection is within content area
        if (!contentArea || !contentArea.contains(range.commonAncestorContainer)) {
          tooltip.classList.remove('show')
          return
        }
        
        selectedText = selection.toString().trim()
        if (selectedText.length < 3) {
          tooltip.classList.remove('show')
          return
        }
        
        tooltipJustShown = true
        self.showAITooltip(selection)
        
        // Reset flag after a short delay
        setTimeout(function() {
          tooltipJustShown = false
        }, 100)
      })
      
      // Handle AI option selection
      document.addEventListener('click', function(e) {
        if (e.target.closest('.ai-tooltip-option')) {
          var action = e.target.closest('.ai-tooltip-option').getAttribute('data-action')
          if (selectedText) {
            self.handleAIAction(action, selectedText)
            var tooltip = document.getElementById('ai-tooltip')
            if (tooltip) tooltip.classList.remove('show')
            window.getSelection().removeAllRanges()
            selectedText = ''
          }
        }
      })
      
      // Hide tooltip on scroll or click outside
      document.addEventListener('scroll', function() {
        var tooltip = document.getElementById('ai-tooltip')
        if (tooltip) tooltip.classList.remove('show')
      }, true)
      
      document.addEventListener('click', function(e) {
        // Don't hide tooltip if it was just shown
        if (tooltipJustShown) return
        
        var tooltip = document.getElementById('ai-tooltip')
        if (tooltip && !tooltip.contains(e.target)) {
          tooltip.classList.remove('show')
        }
      })
    },
    
    showAITooltip: function(selection) {
      var tooltip = document.getElementById('ai-tooltip')
      if (!tooltip) return
      
      var rect = selection.getRangeAt(0).getBoundingClientRect()
      
      // Show tooltip to calculate its dimensions
      tooltip.classList.add('show')
      
      // Wait for next frame to get accurate dimensions
      requestAnimationFrame(function() {
        var tooltipWidth = tooltip.offsetWidth || 220
        var tooltipHeight = tooltip.offsetHeight + 12 // Include arrow height
        
        // Position tooltip above the selection, centered
        var left = Math.max(10, Math.min(
          window.innerWidth - tooltipWidth - 10,
          rect.left + rect.width / 2 - tooltipWidth / 2
        ))
        var top = rect.top - tooltipHeight - 5
        
        // If tooltip would be above viewport, show it below the selection
        if (top < 10) {
          top = rect.bottom + 10
        }
        
        tooltip.style.left = left + 'px'
        tooltip.style.top = top + 'px'
      })
    },
    
    handleAIAction: function(action, text) {
      // Set the context text
      this.chatContext = text
      
      // Open chat panel if not already open
      if (!this.chatPanelVisible) {
        this.chatPanelVisible = true
      }
      
      var prompt = ''
      var autoSend = false
      
      switch (action) {
        case 'explain':
          prompt = this.aiExplainPrompt || 'Please explain this'
          autoSend = true
          break
        case 'summarize':
          prompt = this.aiSummarizePrompt || 'Please summarize this'
          autoSend = true
          break
        case 'question':
          prompt = ''  // Empty prompt for custom questions
          autoSend = false
          break
      }
      
      this.chatInput = prompt
      
      if (autoSend && prompt) {
        // Auto-send for explain and summarize
        var self = this
        this.$nextTick(function() {
          setTimeout(function() {
            self.sendChatMessage()
          }, 50)
        })
      } else {
        // Focus the chat input for question (or when no prompt)
        var self = this
        this.$nextTick(function() {
          setTimeout(function() {
            var chatInput = document.querySelector('#chat-panel input[type="text"]')
            if (chatInput) chatInput.focus()
          }, 100)
        })
      }
    },
    sendChatMessageWithAction: function(action) {
      if (!this.chatInput.trim() || this.loading.chat || !this.itemSelectedDetails) return
      
      var messageContent = this.chatInput.trim()
      
      // Store the message for display (showing the selected text)
      this.chatMessages.push({
        role: 'user',
        content: messageContent,
        hasContext: true,
        context: messageContent // The selected text is the context
      })
      
      // For API call, send the selected text as a user message
      var messages = [{
        role: 'user',
        content: messageContent
      }]
      
      this.chatInput = ''
      this.chatContext = '' // Clear context after sending
      this.loading.chat = true
      
      var item = this.itemSelectedDetails
      var content = this.itemSelectedReadability || item.content || ''
      var vm = this
      
      this.$nextTick(function() {
        var chatContainer = vm.$refs.chatContainer
        if (chatContainer) {
          chatContainer.scrollTop = chatContainer.scrollHeight
        }
      })
      
      api.chat(messages, item.title, content, action).then(function(data) {
        vm.loading.chat = false
        if (data.error) {
          vm.chatMessages.push({
            role: 'assistant',
            content: 'Sorry, I encountered an error: ' + data.error
          })
        } else {
          vm.chatMessages.push({
            role: 'assistant',
            content: data.response
          })
        }
        vm.$nextTick(function() {
          var chatContainer = vm.$refs.chatContainer
          if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight
          }
        })
      }).catch(function(error) {
        vm.loading.chat = false
        vm.chatMessages.push({
          role: 'assistant',
          content: 'Sorry, I encountered an error: ' + error.message
        })
        vm.$nextTick(function() {
          var chatContainer = vm.$refs.chatContainer
          if (chatContainer) {
            chatContainer.scrollTop = chatContainer.scrollHeight
          }
        })
      })
    },
    
    clearChatContext: function() {
      this.chatContext = ''
    },
  }
})

vm.$mount('#app')
