#include "hostsmodel.h"

#include <QtCore/QSettings>
#include <QtCore/QUrl>

#include "serverapi.h"

HostsModel::HostsModel(QObject *parent, QSettings *settings, ServerApi *api)
    : QStandardItemModel(parent),
      mSettings(settings),
      mApi(api)
{
    QStringList hosts;

    settings->beginGroup("hosts");

    const int size = settings->beginReadArray("host");
    for (int i = 0; i < size; i++) {
        settings->setArrayIndex(i);
        const QString addr = settings->value("address").toString();
        if (QUrl(addr).isValid()) {
            hosts.append(addr);
        }
    }
    settings->endArray();

    settings->endGroup();

    foreach (QString addr, hosts) {
        QStandardItem *item = new QStandardItem(addr);
        appendRow(item);
    }

    QString defaultHost = getDefaultHost();
    if (!defaultHost.isEmpty()) {
        api->setHost(defaultHost);
    }
}

bool HostsModel::addHost(const QString &addr)
{
    if (!QUrl(addr).isValid()) {
        return false;
    }

    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *it = item(row);
        if (it->data(Qt::DisplayRole).toString() == addr) {
            return true;
        }
    }

    QStandardItem *newItem = new QStandardItem(addr);
    appendRow(newItem);

    mSettings->beginGroup("hosts");
    mSettings->beginWriteArray("host");
    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *it = item(row);
        mSettings->setArrayIndex(row);
        mSettings->setValue("address", it->data(Qt::DisplayRole));
    }
    mSettings->endArray();
    mSettings->endGroup();

    return true;
}

bool HostsModel::setDefaultHost(const QString &addr)
{
    if (!addHost(addr)) {
        return false;
    }

    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *it = item(row);
        if (it->data(Qt::DisplayRole).toString() == addr) {
            it->setData(true);
            mApi->setHost(addr);
            return true;
        }
    }

    return false;
}

QString HostsModel::getDefaultHost() const
{
    for (int row = 0; row < rowCount(); row++) {
        QStandardItem *it = item(row);
        if (it->data().toBool() == true) {
            return it->data(Qt::DisplayRole).toString();
        }
    }

    return "";
}
