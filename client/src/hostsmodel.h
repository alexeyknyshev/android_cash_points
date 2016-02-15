#ifndef HOSTSMODEL_H
#define HOSTSMODEL_H

#include <QtGui/QStandardItemModel>

class QSettings;
class ServerApi;

class HostsModel : public QStandardItemModel
{
public:
    HostsModel(QObject *parent, QSettings *settings, ServerApi *api);

    Q_INVOKABLE bool addHost(const QString &addr);
    Q_INVOKABLE bool setDefaultHost(const QString &addr);
    Q_INVOKABLE QString getDefaultHost() const;

private:
    QSettings *mSettings;
    ServerApi *mApi;
};

#endif // HOSTSMODEL_H
