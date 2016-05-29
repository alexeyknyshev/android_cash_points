#ifndef GOOGLEAPICLIENT_H
#define GOOGLEAPICLIENT_H

#include <QtCore/QObject>

class GoogleApiClient : public QObject
{
    Q_OBJECT

public:
    GoogleApiClient(QObject *parent);

public slots:
    void dial();
};

#endif // GOOGLEAPICLIENT_H
