var component

function newLoader(callbackFunc, act) {
    if (!act) {
        console.assert(act, "act is not defined!")
        return
    }

    if (!component) {
        component = Qt.createComponent("ViewLoader.qml")
    }

    var self = {
        "action": act,
        "callback": callbackFunc,
        "component": component,
    }

    if (self.component.status === Component.Ready) {
        finishCreation(self);
    } else {
        component.statusChanged.connect(function() { finishCreation(self) });
    }

    return self
}

function finishCreation(self) {
    if (self.component.status === Component.Ready) {
        var object

        if (!self.action.properties) {
            self.action.properties = {}
        }

        if (self.action.component) {
            var count = self.action.component.children.length
            for (var i = 0; i < count; ++i) {
                self.action.component.children[i].destroy()
            }
            self.action.properties["anchors.fill"] = self.action.component
            object = self.component.createObject(self.action.component, self.action.properties)
        } else {
            object = self.component.createObject(null, self.action.properties);
        }

        if (!object) {
            console.log("Error creating ViewLoader object");
        }
        if (self.callback) {
            self.callback(object)
        }
    } else if (self.component.status === Component.Error) {
        console.log("Error loading component:", self.component.errorString());
    }
}
