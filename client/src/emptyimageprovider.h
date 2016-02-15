#ifndef EMPTYIMAGEPROVIDER_H
#define EMPTYIMAGEPROVIDER_H

#include <QtQuick/QQuickImageProvider>

class EmptyImageProvider : public QQuickImageProvider
{
public:
    EmptyImageProvider();

    virtual QImage requestImage(const QString &, QSize *size, const QSize &requestedSize) override;
};

#endif // NUMBERIMAGEPROVIDER_H
