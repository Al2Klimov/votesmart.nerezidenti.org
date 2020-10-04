import React, {Component, ReactNode} from 'react';

import {
  Button,
  SafeAreaView,
  StyleSheet,
  ScrollView,
  View,
  Text,
  StatusBar,
} from 'react-native';

import {Colors} from 'react-native/Libraries/NewAppScreen';

export default class App extends Component<{}, {view: 'is-russian'}> {
  render(): ReactNode {
    const sections: ReactNode[] = [];

    if (this.state === null) {
      this.setState({view: 'is-russian'});
    } else {
      switch (this.state.view) {
        case 'is-russian':
          sections.push(
            <>
              <View style={styles.sectionContainer}>
                <Text style={styles.sectionTitle}>Вы –</Text>
                <Text style={styles.sectionDescription}>
                  гражданин Российской Федерации?
                </Text>
                <Text style={styles.sectionDescription} />
                <Button title="Да" onPress={() => {}} />
                <Text style={styles.sectionDescription} />
                <Button title="Нет" onPress={() => {}} />
              </View>
            </>,
          );
      }
    }

    return (
      <>
        <StatusBar barStyle="dark-content" />
        <SafeAreaView>
          <ScrollView
            contentInsetAdjustmentBehavior="automatic"
            style={styles.scrollView}>
            <View style={styles.body}>
              <View style={styles.sectionContainer}>
                <Text style={styles.sectionTitle}>Добро пожаловать</Text>
                <Text style={styles.sectionDescription}>
                  в умное голосование для нерезидентов!
                </Text>
              </View>
              {sections}
            </View>
          </ScrollView>
        </SafeAreaView>
      </>
    );
  }
}

const styles = StyleSheet.create({
  scrollView: {
    backgroundColor: Colors.lighter,
  },
  body: {
    backgroundColor: Colors.white,
  },
  sectionContainer: {
    marginTop: 32,
    paddingHorizontal: 24,
  },
  sectionTitle: {
    textAlign: 'center',
    fontSize: 24,
    fontWeight: '600',
    color: Colors.black,
  },
  sectionDescription: {
    textAlign: 'center',
    marginTop: 8,
    fontSize: 18,
    fontWeight: '400',
    color: Colors.dark,
  },
});
