import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Controls.Styles 1.4
import QtQuick.Layouts 1.1
import QtQuick.Dialogs 1.2
import QtGraphicalEffects 1.0

import "currency.js" as Currency

Rectangle {
    id: topRect

    MouseArea {
        anchors.fill: parent
        preventStealing: true
    }

    property bool editingMode: false
    property bool existingCashpoint: false
    property var externalAction

    onExternalActionChanged: {
        if (externalAction.data) {
            setCashpointData(externalAction.data)
        }
    }

    property var oldData: ({})
    property var newData: ({})

    function editCurrency(currency, add) {
        if (add) {
            if (!newData.currency) {
                newData.currency = [currency]
                dataChanged()
                return
            }
            if (newData.currency.indexOf(currency) === -1) {
                newData.currency.push(currency)
                dataChanged()
            }
        } else {
            var index = newData.currency.indexOf(currency)
            if (index > -1) {
                newData.currency.splice(index, 1)
                if (newData.currency.length === 0) {
                    delete newData.currency
                }
                dataChanged()
            }
        }
    }

    signal finished(var result)

    function dataChanged() {
        var patch = getPatchObject()
        if (acceptButton) {
            if (Object.keys(patch).length > 0) {
                acceptButton.state = "edited"
            } else {
                acceptButton.state = ""
            }
        }

        if (bankIco) {
            if (patch.bank_id) {
                bankIco.source = "image://ico/bank/" + patch.bank_id
            } else if (oldData.bank_id) {
                bankIco.source = "image://ico/bank/" + oldData.bank_id
            } else {
                bankIco.source = ""
            }
        }
    }

    signal action(var act)

    onAction: {
        if (externalAction && externalAction.actionCallback) {
            externalAction.actionCallback(act)
        }
    }

    function getPatchObject() {
        var patch = {}
        for (var k in newData) {
            if (newData[k] !== oldData[k]) {
                patch[k] = newData[k]
            }
        }
        return patch
    }

    function getFullObject() {
        var longitude = oldData.longitude
        var latitude = oldData.latitude
        var mainOffice = false // TODO: main_office
        var cashIn = false // TODO: cash_in
        var schedule = {} // TODO: schedule
        var tel = "" // TODO: tel
        var additional = "" // TODO: additional

        var currency = []
        if (rubCheckBox.checked) {
            currency.append(Currency.RUB)
        }
        if (usdCheckBox.checked) {
            currency.append(Currency.USD)
        }
        if (eurCheckBox.checked) {
            currency.append(Currency.EUR)
        }

        return {
            "type": typeComboBox.getTypeString(),
            "bank_id": bankSelector.bankId,
            "town_id": townSelector.townId,
            "longitude": longitude,
            "latitude": latitude,
            "address_comment": addressTextEdit.text,
            "main_office": mainOffice,
            "free_access": freeAccessCheckBox.checked,
            "without_weekend": withoutWeekendCheckBox.checked,
            "works_as_shop": worksAsShopCheckBox.checked,
            "currency": currency,
            "cash_in": cashIn,
            "round_the_clock": roundTheClockCheckBox.checked,
            "schedule": schedule,
            "tel": tel,
            "additional": additional,
        }
    }

    function resetCashpointData() {
        _setCashpointData(oldData)
        dataChanged()
    }

    function setCashpointData(opt) {
        if (!opt.action) {
            return false
        }
        if (!opt.data) {
            return false
        }
        if (!opt.data.longitude || !opt.data.latitude) {
            return false
        }

        var eMode = false // editing mode disabled
        if (opt.action === "create" || opt.action === "edit") {
            eMode = true
        }

        editingMode = eMode

        _setCashpointData(opt.data)
    }

    function _setCashpointData(data) {
        newData = {}
        acceptButton.state = ""

        existingCashpoint = idLabel.setId(data.id)

        oldData = data
        typeComboBox.setType(data.type)
        bankSelector.setBankId(data.bank_id)
        townSelector.setTownId(data.town_id)
        addressEdit.text = "" // TODO: addressComment
        freeAccessCheckBox.set(data.free_access)
        withoutWeekendCheckBox.set(data.without_weekend)
        roundTheClockCheckBox.set(data.round_the_clock)
        rubCheckBox.set(data.currency.indexOf(Currency.RUB) > -1)
        usdCheckBox.set(data.currency.indexOf(Currency.USD) > -1)
        eurCheckBox.set(data.currency.indexOf(Currency.EUR) > -1)
    }

    Rectangle {
        id: topMenu
        z: parent.z + 2
        anchors.top: parent.top
        anchors.left: parent.left
        anchors.right: parent.right
        height: Math.max(parent.height, parent.width) * 0.08 *
                (topRect.existingCashpoint ? 1 : 0)

        color: "steelblue"
        visible: topRect.existingCashpoint

        MouseArea {
            anchors.fill: parent
        }

        Image {
            anchors.top: parent.top
            anchors.right: switchEditingMode.left
            anchors.bottom: parent.bottom
            anchors.margins: parent.height * 0.2

            width: height

            sourceSize: Qt.size(width, height)
            source: topRect.editingMode ? "image://ico/share.svg" : "image://ico/share.svg"

            MouseArea {
                id: shareButtonMouseArea
                anchors.fill: parent
                onClicked: {

                }
            }

            scale: shareButtonMouseArea.pressed ? 0.9 : 1.0
        }

        Image {
            id: switchEditingMode
            anchors.top: parent.top
            anchors.right: parent.right
            anchors.bottom: parent.bottom
            anchors.margins: parent.height * 0.2

            width: height

            sourceSize: Qt.size(width, height)
            source: topRect.editingMode ? "image://ico/eye.svg" : "image://ico/editing.svg"

            MouseArea {
                id: switchEditingModeMouseArea
                anchors.fill: parent
                onClicked: {
                    if (topRect.editingMode) {

                        topRect.editingMode = false
                    } else {
                        topRect.editingMode = true
                    }
                }
            }

            MessageDialog {
                id: saveBeforeClosingDialog
                title: qsTr("")
            }

            scale: switchEditingModeMouseArea.pressed ? 0.9 : 1
        }

        Image {
            id: refreshButtonIco
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.bottom: parent.bottom
            anchors.margins: parent.height * 0.2

            width: height
            sourceSize: Qt.size(width, height)
            source: "image://ico/clear.svg"

            opacity: topRect.editingMode ? 1 : 0

            Behavior on opacity {
                NumberAnimation {
                    duration: 200
                }
            }

            MouseArea {
                id: refreshButtonIcoMouseArea
                anchors.fill: parent
                onClicked: {
                    topRect.resetCashpointData()
                }
            }

            scale: refreshButtonIcoMouseArea.pressed ? 0.9 : 1
        }
    }

    RectangularGlow {
        z: topMenu.z - 1
        visible: topMenu.visible
        anchors.fill: topMenu
        glowRadius: topMenu.height / 3
        spread: 0.2
        color: "#11000055"
        cornerRadius: glowRadius
    }

    Flickable {
        id: flickable
        anchors.top: topMenu.bottom
        anchors.left: parent.left
        anchors.bottom: parent.bottom
        anchors.right: parent.right
        anchors.margins: Math.min(parent.width, parent.height) * 0.03
        flickableDirection: Flickable.VerticalFlick
        contentWidth: grid.width
        contentHeight: grid.height

        Rectangle {
            id: grid
            height: acceptButton.y + acceptButton.height // flickable.height * 1.4
            width: flickable.width

            Label {
                anchors.top: parent.top
                anchors.left: parent.left
                height: flickable.height * 0.1

                id: idLabel_
                text: qsTr("Id метки")
                visible: existingCashpoint
            }

            Label {
                anchors.top: parent.top
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: idLabel
                visible: existingCashpoint

                function setId(id) {
                    if (id && id > 0) {
                        text = id.toString()
                        return true
                    }

                    text = ""
                    return false
                }
            }

            Label {
                anchors.top: idLabel.bottom
                anchors.left: parent.left
                height: flickable.height * 0.12

                id: typeLabel_

                text: qsTr("Тип")
                verticalAlignment: Qt.AlignVCenter
            }

            Label {
                anchors.top: idLabel_.bottom
                anchors.right: idLabel.right
                height: typeLabel_.height

                id: typeLabel
                visible: !editingMode
                text: typeComboBox.currentText

                verticalAlignment: Qt.AlignVCenter
                horizontalAlignment: Qt.AlignRight
            }

            ComboBox {
                anchors.top: idLabel_.bottom
                anchors.right: idLabel.right
                height: flickable.height * 0.1
                width: flickable.width * 0.5

                id: typeComboBox
                visible: editingMode
                currentIndex: 0
                model: [ qsTr("Банкомат"), qsTr("Офис"), qsTr("Касса") ]

                function getTypeString() {
                    switch (currentIndex) {
                        case 0: return "atm"
                        case 1: return "office"
                        case 2: return "cash"
                    }
                    return ""
                }

                onCurrentIndexChanged: {
                    var type = getTypeString()
                    if (type.length > 0) {
                        topRect.newData.type = type
                        dataChanged()
                    }
                }

                function setType(type) {
                    if (!type) {
                        type = "atm"
                    }

                    if (type === "atm") {
                        currentIndex = 0
                    } else if (type === "office" || type === "branch") {
                        currentIndex = 1
                    } else if (type === "cash") {
                        currentIndex = 2
                    } else {
                        currentIndex = 0
                    }
                }
            }

            Label {
                anchors.top: typeLabel_.bottom
                anchors.left: parent.left
                height: flickable.height * 0.12

                id: bankLabel_

                text: qsTr("Банк")
                verticalAlignment: Qt.AlignVCenter
            }

            Image {
                anchors.top: bankLabel.top
                anchors.right: bankLabel.left
                anchors.bottom: bankLabel.bottom
                anchors.margins: bankLabel.height * 0.25

                id: bankIco

                visible: !editingMode
                width: height
                sourceSize: Qt.size(width, height)
            }

            Label {
                anchors.top: typeLabel_.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: bankLabel

                visible: !editingMode
                text: bankSelector.text
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
                horizontalAlignment: Qt.AlignRight
                verticalAlignment: Qt.AlignVCenter
            }

            Button {
                anchors.top: typeLabel_.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1
                width: flickable.width * 0.5

                id: bankSelector
                visible: editingMode

                text: defaultCaption
                property string defaultCaption: qsTr("Выбрать банк")
                property int bankId: 0

                style: ButtonStyle {
                    label: Label {
                        anchors.fill: parent
                        text: bankSelector.text
                        verticalAlignment: Text.AlignVCenter
                        horizontalAlignment: Text.AlignHCenter
                        wrapMode: Text.WordWrap
                    }
                }

                onClicked: {
                    if (topRect.editingMode) {
                        topRect.action({
                            "type": "openView",
                            "path": "BanksList.qml",
                            "callback": function(result) {
                                bankId = result.id
                                topRect.action({ "type": "undo" })
                            },
                        })
                    }
                }

                onBankIdChanged: {
                    topRect.newData.bank_id = bankId
                    if (bankId == 0) {
                        text = defaultCaption
                        dataChanged()
                    } else {
                        var bankJsonData = bankListModel.getBankData(bankId)
                        if (bankJsonData) {
                            var bank = JSON.parse(bankJsonData)
                            text = bank.name
                            dataChanged()
                        } else {
                            bankId = 0
                        }
                    }
                }

                function setBankId(id) {
                    bankId = id ? id : 0
                }
            }

            Label {
                anchors.top: bankLabel_.bottom
                anchors.left: parent.left
                height: flickable.height * 0.12

                id: townLabel_

                text: qsTr("Населённый\nпункт")
                verticalAlignment: Qt.AlignVCenter
            }

            Label {
                anchors.top: bankLabel_.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: townLabel

                visible: !editingMode
                text: townSelector.text
                wrapMode: Text.WrapAtWordBoundaryOrAnywhere
                horizontalAlignment: Qt.AlignRight
                verticalAlignment: Qt.AlignVCenter
            }

            Button {
                anchors.top: bankLabel_.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1
                width: flickable.width * 0.5

                id: townSelector
                visible: editingMode

                style: ButtonStyle {
                    id: style
                    label: Label {
                        anchors.fill: parent
                        text: townSelector.text
                        verticalAlignment: Text.AlignVCenter
                        horizontalAlignment: Text.AlignHCenter
                        wrapMode: Text.WordWrap
                    }
                }

                text: defaultCaption
                property string defaultCaption: qsTr("Выбрать населённый пункт")
                property int townId: 0

                onClicked: {
                    if (topRect.editingMode) {
                        topRect.action({
                            "type": "openView",
                            "path": "TownList.qml",
                            "callback": function(result) {
                                townId = result.id
                                topRect.action({ "type": "undo" })
                            },
                        })
                    }
                }

                onTownIdChanged: {
                    topRect.newData.town_id = townId
                    if (townId == 0) {
                        text = defaultCaption
                        dataChanged()
                    } else {
                        var townJsonData = townListModel.getTownData(townId)
                        if (townJsonData) {
                            var town = JSON.parse(townJsonData)
                            text = town.name
                            dataChanged()
                        } else {
                            townId = 0
                        }
                    }
                }

                function setTownId(id) {
                    townId = id ? id : 0
                }
            }

            Label {
                anchors.top: townLabel_.bottom
                anchors.left: parent.left
                height: flickable.height * 0.12

                id: addressLabel

                text: qsTr("Адрес (комментарий)")
            }

            TextEdit {
                anchors.top: townLabel_.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1
                width: flickable.width * 0.5

                id: addressEdit

                wrapMode: TextEdit.WordWrap
                readOnly: !editingMode

                function setAddressComment(comment) {
                    if (!comment) {
                        comment = ""
                    }

                    text = comment
                }

                onTextChanged: {
                    topRect.newData.address_comment = text
                    dataChanged()
                }
            }

            Label {
                anchors.top: addressEdit.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("В свободном доступе")
            }

            CheckBox {
                anchors.top: addressEdit.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: freeAccessCheckBox

                enabled: editingMode

                function set(freeAccess) {
                    checked = freeAccess ? freeAccess : false
                }

                onClicked: {
                    topRect.newData["free_access"] = checked
                    dataChanged()
                }
            }

            Label {
                anchors.top: freeAccessCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Без выходных")
            }

            CheckBox {
                anchors.top: freeAccessCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: withoutWeekendCheckBox
                enabled: editingMode

                function set(withoutWeekend) {
                    checked = withoutWeekend ? withoutWeekend : false
                }

                onClicked: {
                    topRect.newData.without_weekend = checked
                    dataChanged()
                }
            }

            Label {
                anchors.top: withoutWeekendCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Работает в режиме\nточки установки")
            }

            CheckBox {
                anchors.top: withoutWeekendCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: worksAsShopCheckBox
                enabled: editingMode

                onClicked: {
                    topRect.newData.works_as_shop = checked
                    dataChanged()
                }
            }

            Label {
                anchors.top: worksAsShopCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Круглосуточно")
            }

            CheckBox {
                anchors.top: worksAsShopCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: roundTheClockCheckBox
                enabled: editingMode

                function set(roundTheClock) {
                    checked = roundTheClock ? true : false
                }

                onClicked: {
                    topRect.newData.round_the_clock = checked
                    dataChanged()
                }
            }

            Label {
                anchors.top: roundTheClockCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Рубли")
            }

            CheckBox {
                anchors.top: roundTheClockCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: rubCheckBox
                enabled: editingMode

                function set(rub) {
                    checked = rub ? rub : false
                }

                onClicked: {
                    editCurrency(Currency.RUB, checked)
                }
            }

            Label {
                anchors.top: rubCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Доллары")
            }

            CheckBox {
                anchors.top: rubCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: usdCheckBox
                enabled: editingMode

                function set(usd) {
                    checked = usd ? usd : false
                }

                onClicked: {
                    editCurrency(Currency.USD, checked)
                }
            }

            Label {
                anchors.top: usdCheckBox.bottom
                anchors.left: parent.left
                height: flickable.height * 0.1

                verticalAlignment: Qt.AlignVCenter

                text: qsTr("Евро")
            }

            CheckBox {
                anchors.top: usdCheckBox.bottom
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: eurCheckBox
                enabled: editingMode

                function set(eur) {
                    checked = eur ? eur : false
                }

                onClicked: {
                    editCurrency(Currency.EUR, checked)
                }
            }

            Button {
                anchors.top: eurCheckBox.bottom
                anchors.left: parent.left
                anchors.right: parent.right
                height: flickable.height * 0.1

                id: acceptButton
                text: qsTr("Закрыть")

                states: [
                    State {
                        name: "edited"
                        PropertyChanges {
                            target: acceptButton
                            text: qsTr("Принять")
                        }
                    }
                ]

                onClicked: {
                    if (state == "edited") {
                        if (bankSelector.bankId == 0) {
                            errDialog.text = qsTr("Пожалуйста, выберите банк")
                            errDialog.open()
                            return
                        }
                        if (townSelector.townId == 0) {
                            errDialog.text = qsTr("Пожалуйста, выберите населённый пункт")
                            errDialog.open()
                            return
                        }
                        if (!rubCheckBox.checked && !usdCheckBox.checked && !eurCheckBox.checked) {
                            errDialog.text = qsTr("Пожалуйста, выберите хотя бы одну валюту")
                            errDialog.open()
                            return
                        }

                        if (editingMode) {
                            if (existingCashpoint) {
                                responseWaitDialog.open()
                                var patch = getPatchObject()
                                patch["id"] = oldData.id
                                cashpointModel.editCashPoint(JSON.stringify(patch), function(id, step, ok, msg) {
                                    console.log("Cashpoint editing:", ok, "(", msg, ")")
                                    responseWaitDialog.close()
                                    finished({ "ok": ok, "msg": msg })
                                })
                            } else {
                                responseWaitDialog.open()
                                cashpointModel.createCashPoint(JSON.stringify(getFullObject()), function(id, step, ok, msg) {
                                    console.log("Cashpoint creating:", ok, "(", msg, ")")
                                    responseWaitDialog.close()
                                    finished({ "ok": ok, "msg": msg })
                                })
                            }
                        }
                    } else if (state == "") {
                        finished({ "ok": true })
                    }
                }
            }
        }
    }

    Dialog {
        id: responseWaitDialog
        contentItem: Rectangle {
            implicitWidth: Math.min(topRect.width, topRect.height) * 0.5
            implicitHeight: implicitWidth

            BusyIndicator {
                anchors.centerIn: parent
            }
        }
    }

    MessageDialog {
        id: errDialog
        standardButtons: StandardButton.Ok
    }

    onFinished: {
        if (result.msg && result.msg.length > 0) {
            errDialog.setText(result.msg)
            errDialog.icon = result.ok ? StandardIcon.Information : StandardIcon.Critical
            errDialog.open()
        }
        topRect.action({ "type": "undo" })
    }
}
