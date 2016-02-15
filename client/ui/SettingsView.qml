import QtQuick 2.0
import QtQuick.Controls 1.4

Rectangle {
    id: topRect

    Column {
        Row {
            CheckBox {

            }

            Label {
                text: qsTr("Save last view position")
            }
        }

        Row {
            ComboBox {
                model: ListModel {
                    id: hostsComboBox
                }
            }
        }
    }
}

