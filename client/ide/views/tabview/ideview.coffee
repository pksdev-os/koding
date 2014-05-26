 class IDEView extends WorkspaceTabView

  constructor: (options = {}, data) ->

    super options, data

    @openFiles = []

    @on 'PlusHandleClicked',   @bound 'createPlusContextMenu'
    @on 'FileNeedsToBeOpened', @bound 'openFile'
    @forwardEvent @tabView, 'FileNeedsToBeOpened'

    @once 'viewAppended', => KD.utils.wait 300, => @createEditor()

  getPlusMenuItems: ->
    return {
      'Editor'        : callback : => @createEditor()
      'Terminal'      : callback : => @createTerminal()
      'Browser'       : callback : => @createPreview()
      'Drawing Board' : callback : => @createDrawingBoard()
    }

  createPane_: (view, paneOptions, paneData) ->
    unless view or paneOptions
      return new Error 'Missing argument for createPane_ helper'

    unless view instanceof KDView
      return new Error 'View must be an instance of KDView'

    pane = new KDTabPaneView paneOptions, paneData
    pane.addSubView view
    @tabView.addPane pane

    pane.once 'KDObjectWillBeDestroyed', => @handlePaneRemoved pane

  createEditor: (file, content) ->
    file        = file    or FSHelper.createFileFromPath @getDummyFilePath()
    content     = content or ''
    editor      = new EditorPane { file, content, delegate: this }
    paneOptions =
      name      : file.name
      editor    : editor

    @createPane_ editor, paneOptions, file

  createTerminal: ->
    @createPane_ new TerminalPane, { name: 'Terminal' }

  createDrawingBoard: ->
    @createPane_ new DrawingPane,  { name: 'Drawing'  }

  createPreview: ->
    @createPane_ new PreviewPane,  { name: 'Browser'  }

  removeOpenDocument: ->
    # TODO: This method is legacy, should be reimplemented in ace bundle.

  click: ->
    super
    KD.getSingleton('appManager').tell 'IDE', 'setActiveTabView', this.tabView

  openFile: (file, content) ->
    if @openFiles.indexOf(file) > -1
      @switchToEditorTabByFile file
    else
      @createEditor file, content
      @openFiles.push file

  switchToEditorTabByFile: (file) ->
    for pane, index in @tabView.panes when file is pane.getData()
      @tabView.showPaneByIndex index

  handlePaneRemoved: (pane) ->
    file = pane.getData()
    @openFiles.splice @openFiles.indexOf(file), 1

  getDummyFilePath: ->
    return 'localfile://Untitled.txt'

  createPlusContextMenu: ->
    offset        = @holderView.plusHandle.$().offset()
    contextMenu   = new KDContextMenu
      delegate    : this
      x           : offset.left - 125
      y           : offset.top  + 30
      arrow       :
        placement : 'top'
        margin    : -20
    , @getPlusMenuItems()

    contextMenu.once 'ContextMenuItemReceivedClick', ->
      contextMenu.destroy()
