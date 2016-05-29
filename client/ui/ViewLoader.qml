import QtQuick 2.0

Rectangle {
    id: topView
    anchors.fill: parent

    signal action(var action)

    function setView(qmlfile, action) {
        if (loader.source !== qmlfile) {
            loader.source = qmlfile
            loader.action = action
        } else {
            loader.initItem(action)
        }
    }

    Loader {
        anchors.fill: parent
        id: loader
        asynchronous: true

        property var action

        function initItem(act) {
            if (item) {
                item.externalAction = act
                if (item.action) {
                    item.action.connect(dispatchAction)
                }
                if (act.initCallback) {
                    act.initCallback(item)
                }
            }

        }

        onLoaded: {
            initItem(action)
        }

        function dispatchAction(action) {
            topView.action(action)
        }

        onStatusChanged: {
            var text = ""
            if (status === Loader.Null) {
                text = "null"
            } else if (status === Loader.Ready) {
                text = "ready"
            } else if (status === Loader.Loading) {
                text = "loading"
            } else if (status === Loader.Error) {
                text = "error"
            }

            console.log("Loader state: " + text)
        }
    }
}

