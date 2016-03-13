#ifndef APPSTATEPROXY_H
#define APPSTATEPROXY_H

#include <QtCore/QObject>

class QGuiApplication;

class AppStateProxy : public QObject
{
    Q_OBJECT

public:
    AppStateProxy(QGuiApplication *app);

signals:
    void serverDataLoaded(bool ok, QString dataType);
    void appStateChanged(int state);

public slots:
    void onConnectionFailed();
    void onTownsDataLoaded();
    void onBanksDataLoaded();

private slots:
    void onAppStateChanged(Qt::ApplicationState state);
};

#endif // APPSTATEPROXY_H
