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
    void appStateChanged(int state);

private slots:
    void onAppStateChanged(Qt::ApplicationState state);
};

#endif // APPSTATEPROXY_H
